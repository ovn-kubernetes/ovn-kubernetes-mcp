#!/usr/bin/env bash

set -exo pipefail

NVM_VERSION=$1
NODE_VERSION=$2
NPM_VERSION=$3
GINKGO_VERSION=$4
MCP_MODE=$5
FOCUS=${6:-""}

if [[ -z "${NVM_VERSION}" ]] || [[ -z "${NODE_VERSION}" ]] || [[ -z "${NPM_VERSION}" ]] || [[ -z "${GINKGO_VERSION}" ]]; then
    echo "NVM_VERSION, NODE_VERSION, NPM_VERSION and GINKGO_VERSION are required"
    exit 1
fi

if [[ "${MCP_MODE}" != "offline" && "${MCP_MODE}" != "live-cluster" && -n "${MCP_MODE}" ]]; then
    echo "Invalid MCP_MODE: ${MCP_MODE}. Must be 'offline' or 'live-cluster'"
    exit 1
fi

install_dependencies() {
    # Install ginkgo
    go install github.com/onsi/ginkgo/v2/ginkgo@"${GINKGO_VERSION}"

    # Install node version manager
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v"${NVM_VERSION}"/install.sh | bash

    export NVM_DIR="${HOME}/.nvm"
    [ -s "${NVM_DIR}/nvm.sh" ] && \. "${NVM_DIR}/nvm.sh"  # This loads nvm
    [ -s "${NVM_DIR}/bash_completion" ] && \. "${NVM_DIR}/bash_completion"  # This loads nvm bash_completion

    # Install node version
    nvm install "${NODE_VERSION}"

    # Use node version
    nvm use "${NODE_VERSION}"

    # Install npm
    npm install -g npm@"${NPM_VERSION}" --force

    # Check npm version
    npx -v
}

# install_image accepts the image name along with the tag as an argument and installs it.
install_image() {
  if [ "$CONTAINER_RUNTIME" == "podman" ]; then
    # podman: cf https://github.com/kubernetes-sigs/kind/issues/2027
    rm -f /tmp/image.tar
    podman save -o /tmp/image.tar "${1}"
    kind load image-archive /tmp/image.tar --name "${2}"
  else
    kind load docker-image "${1}" --name "${2}"
  fi
}

install_dependencies
echo "Dependencies installed"

# Run e2e tests
echo "Running e2e tests"
if [[ "${MCP_MODE}" == "offline" ]]; then
    echo "Running offline mode tests (sosreport and must-gather)"
    export MCP_MODE="offline"
    if [[ -n "${FOCUS}" ]]; then
        ginkgo -vv --focus="\[offline\].*${FOCUS}" test/e2e
    else
        ginkgo -vv --focus="\[offline\]" test/e2e
    fi
else
    # Live-cluster: use Dockerfile and k8s manifests (build image, deploy, port-forward)
    if [[ -z "${KUBECTL:-}" ]]; then
        if kubectl_path=$(command -v kubectl 2>/dev/null); then
            KUBECTL=$kubectl_path
        else
            echo "KUBECTL is not set and kubectl was not found in PATH"
            exit 1
        fi
    fi
    export KUBECTL

    if [[ -z "${KUBECONFIG:-}" ]]; then
        echo "KUBECONFIG is required for live-cluster mode"
        exit 1
    fi
    if ! command -v kind &>/dev/null; then
        echo "kind is required for live-cluster mode but was not found. Install it from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
        exit 1
    fi
    GIT_ROOT="${GIT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null)}"
    if [[ -z "${GIT_ROOT:-}" ]]; then
        echo "GIT_ROOT is not set and could not be determined. Run from inside the repo or set GIT_ROOT."
        exit 1
    fi
    cd "$GIT_ROOT" || exit 1

    # Live-cluster e2e targets a kind cluster; infer name from current context (must be kind-<name>).
    CURRENT_CTX=$("${KUBECTL}" config current-context 2>/dev/null || true)
    if [[ -z "${CURRENT_CTX}" ]]; then
        echo "No current kubectl context (KUBECONFIG=${KUBECONFIG}); cannot determine kind cluster name"
        exit 1
    fi
    if [[ "${CURRENT_CTX}" != kind-* ]]; then
        echo "Current kubectl context must be a kind cluster (expected kind-<name>, got: ${CURRENT_CTX})"
        exit 1
    fi
    KIND_CLUSTER_NAME="${CURRENT_CTX#kind-}"
    
    IMAGE="${IMAGE:-localhost/ovnk-mcp-server:dev}"
    MCP_LOCAL_PORT="${MCP_LOCAL_PORT:-18080}"

    DEPLOY_K8S_INVOKED=0
    PF_PID=""
    cleanup() {
        if [[ -n "${PF_PID:-}" ]] && kill -0 "$PF_PID" 2>/dev/null; then
            kill "$PF_PID" 2>/dev/null || true
            wait "$PF_PID" 2>/dev/null || true
        fi
        if [[ "${DEPLOY_K8S_INVOKED}" -eq 1 ]]; then
            make undeploy-k8s || true
        fi
    }
    trap cleanup EXIT

    if [[ -z "${CONTAINER_RUNTIME:-}" ]]; then
        if command -v podman >/dev/null 2>&1; then
            CONTAINER_RUNTIME=podman
        else
            CONTAINER_RUNTIME=docker
        fi
    fi
    export CONTAINER_RUNTIME

    make build-image IMAGE="$IMAGE" CONTAINER_RUNTIME="$CONTAINER_RUNTIME"
    install_image "$IMAGE" "$KIND_CLUSTER_NAME"
    DEPLOY_K8S_INVOKED=1
    make deploy-k8s IMAGE="$IMAGE"

    # Wait for the deployment to exist (and be Available) before querying namespace/svc/port
    MCP_LABEL="app.kubernetes.io/name=ovn-kubernetes-mcp"
    echo "Waiting for MCP server deployment to be Available..."
    if ! "${KUBECTL}" wait --for=condition=Available deployment -A -l "${MCP_LABEL}" --timeout=60s; then
        echo "Deployment with label ${MCP_LABEL} did not become Available within 60s"
        exit 1
    fi

    # Get namespace, service name, and service port from the cluster (what was actually applied)
    NAMESPACE=$("${KUBECTL}" get deployment -A -l "${MCP_LABEL}" -o jsonpath='{.items[0].metadata.namespace}' 2>/dev/null || true)
    if [[ -z "${NAMESPACE:-}" ]]; then
        echo "Could not find deployment with label ${MCP_LABEL}; is deploy-k8s applied?"
        exit 1
    fi
    SVC=$("${KUBECTL}" get svc -n "${NAMESPACE}" -l "${MCP_LABEL}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
    SVC_PORT=$("${KUBECTL}" get svc -n "${NAMESPACE}" -l "${MCP_LABEL}" -o jsonpath='{.items[0].spec.ports[0].port}' 2>/dev/null || true)
    if [[ -z "${SVC:-}" || -z "${SVC_PORT:-}" ]]; then
        echo "Could not find service or port in namespace ${NAMESPACE} with label ${MCP_LABEL}"
        exit 1
    fi

    # Wait for pod to be Ready (readiness probe passed) before port-forward so MCP server accepts requests
    echo "Waiting for MCP server pod to be Ready..."
    if ! "${KUBECTL}" wait --for=condition=Ready pod -n "$NAMESPACE" -l app.kubernetes.io/name=ovn-kubernetes-mcp --timeout=120s; then
        echo "MCP server pod did not become Ready within 120s"
        "${KUBECTL}" get pods -n "$NAMESPACE" -l app.kubernetes.io/name=ovn-kubernetes-mcp 2>/dev/null || true
        exit 1
    fi

    "${KUBECTL}" port-forward -n "$NAMESPACE" "svc/${SVC}" "${MCP_LOCAL_PORT}:${SVC_PORT}" &
    PF_PID=$!
    # Wait for port-forward to be ready: process alive and server responding.
    PF_TIMEOUT=30
    PF_ELAPSED=0
    while [[ $PF_ELAPSED -lt $PF_TIMEOUT ]]; do
        if ! kill -0 "$PF_PID" 2>/dev/null; then
            echo "Port-forward exited unexpectedly"
            exit 1
        fi
        if curl -s -o /dev/null --connect-timeout 2 "http://127.0.0.1:${MCP_LOCAL_PORT}/" 2>/dev/null; then
            break
        fi
        sleep 1
        PF_ELAPSED=$((PF_ELAPSED + 1))
    done
    if [[ $PF_ELAPSED -ge $PF_TIMEOUT ]]; then
        echo "Port-forward did not become ready within ${PF_TIMEOUT}s (http://127.0.0.1:${MCP_LOCAL_PORT} not responding)"
        exit 1
    fi

    export MCP_SERVER_URL="http://127.0.0.1:${MCP_LOCAL_PORT}"
    export MCP_MODE="live-cluster"

    echo "Running live-cluster mode tests (excluding offline tests)"
    if [[ -n "${FOCUS}" ]]; then
        ginkgo -vv --skip="\[offline\]" --focus="${FOCUS}" test/e2e
    else
        ginkgo -vv --skip="\[offline\]" test/e2e
    fi
fi
