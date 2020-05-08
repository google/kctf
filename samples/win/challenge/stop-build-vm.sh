#!/bin/bash

set -Eeuo pipefail
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )"
source "${DIR}/vm_config.sh"

gcloud compute instances delete "${VM_NAME}"
gcloud compute firewall-rules delete "${FIREWALL_RULE}"
rm windows_creds
