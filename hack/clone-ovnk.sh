#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
# SPDX-License-Identifier: Apache-2.0

set -eo pipefail

clone_ovnk() {
    rm -rf "${OVN_KUBERNETES_DIR}"
    git clone https://github.com/ovn-org/ovn-kubernetes.git "${OVN_KUBERNETES_DIR}"
}
