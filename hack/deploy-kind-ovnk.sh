#!/usr/bin/env bash

set -eo pipefail

# Returns the full directory name of the script
DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

source "${DIR}/clone-ovnk.sh"

export OVN_KUBERNETES_DIR="${OVN_KUBERNETES_DIR:-/tmp/ovn-kubernetes}"

clone_ovnk
cd "${OVN_KUBERNETES_DIR}"

export PLATFORM_IPV4_SUPPORT=${PLATFORM_IPV4_SUPPORT:-true}
make -C test install-kind
