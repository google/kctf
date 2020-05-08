#!/bin/bash

set -Eeuo pipefail
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )"
source "${DIR}/vm_config.sh"

echo "[*] creating windows vm"
gcloud beta compute instances create "${VM_NAME}" \
  --image="${IMAGE}" \
  --image-project=windows-cloud \
  --machine-type "${MACHINE_TYPE}" \
  --scopes=cloud-platform,storage-full \
  --metadata windows-startup-script-cmd='winrm set winrm/config/Service/Auth @{Basic="true"}' \
  --tags "${VM_TAG}"

gcloud compute firewall-rules create "${FIREWALL_RULE}" \
  --allow=tcp:5986 \
  --description "${FIREWALL_DESCRIPTION}" \
  --target-tags "${VM_TAG}"

while true; do
  echo -n "[*] Trying to fetch windows credentials, remaining timeout: ${CREDENTIAL_FETCH_TIMEOUT}s"
  gcloud --quiet beta compute reset-windows-password "${VM_NAME}" > windows_creds 2>/dev/null && break
  if [[ ${CREDENTIAL_FETCH_TIMEOUT} -lt ${CREDENTIAL_FETCH_SLEEP} ]]; then
    echo ": failed, giving up"
    exit 1
  fi
  echo ": failed, sleeping ${CREDENTIAL_FETCH_SLEEP}s"
  sleep ${CREDENTIAL_FETCH_SLEEP}
  CREDENTIAL_FETCH_TIMEOUT=$((${CREDENTIAL_FETCH_TIMEOUT} - ${CREDENTIAL_FETCH_SLEEP}))
done
echo ': success'
