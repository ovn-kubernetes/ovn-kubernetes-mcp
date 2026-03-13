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
    echo "Running live-cluster mode tests (excluding offline tests)"
    export MCP_MODE="live-cluster"
    if [[ -n "${FOCUS}" ]]; then
        ginkgo -vv --skip="\[offline\]" --focus="${FOCUS}" test/e2e
    else
        ginkgo -vv --skip="\[offline\]" test/e2e
    fi
fi
