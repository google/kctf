#!/bin/bash

set -Eeuo pipefail
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )"
source "${DIR}/vm_config.sh"

gcloud --quiet beta compute reset-windows-password "${VM_NAME}" > "${DIR}/windows_creds"
