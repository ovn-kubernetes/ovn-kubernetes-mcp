#!/usr/bin/env bash

# set -eo pipefail

NVM_VERSION=$1
NODE_VERSION=$2
NPM_VERSION=$3
MCP_MODE=$4

if [[ -z "${NVM_VERSION}" ]] || [[ -z "${NODE_VERSION}" ]] || [[ -z "${NPM_VERSION}" ]]; then
    echo "NVM_VERSION, NODE_VERSION and NPM_VERSION are required"
    exit 1
fi

install_dependencies() {
    # Install ginkgo
    go install github.com/onsi/ginkgo/v2/ginkgo@latest

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
if [[ "${MCP_MODE}" == "Offline Mode" ]]; then
    echo "Running offline mode tests (sosreport and must-gather)"
    export MCP_MODE="Offline Mode"
    ginkgo -vv --focus="Sosreport Tools" test/e2e
else
    echo "Running online mode tests (excluding offline tests)"
    ginkgo -vv --skip="Sosreport Tools" test/e2e
fi
