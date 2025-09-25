#!/usr/bin/env bash

set -eo pipefail

# Returns the full directory name of the script
DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

source "${DIR}/clone-ovnk.sh"

export OVN_KUBERNETES_DIR="${OVN_KUBERNETES_DIR:-/tmp/ovn-kubernetes}"

clone_ovnk

"${OVN_KUBERNETES_DIR}"/contrib/kind.sh --delete
