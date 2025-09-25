#!/usr/bin/env bash

set -eo pipefail

clone_ovnk() {
    rm -rf "${OVN_KUBERNETES_DIR}"
    git clone https://github.com/ovn-org/ovn-kubernetes.git "${OVN_KUBERNETES_DIR}"
}
