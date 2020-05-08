#!/bin/bash

set -Eeuo pipefail

VM_TAG="win-build-vm"
MACHINE_TYPE="n1-standard-4"
IMAGE="windows-server-1909-dc-core-for-containers-v20200414"
VM_NAME="win-build-vm"
FIREWALL_RULE="gke-kctf-win-build-vm"
FIREWALL_DESCRIPTION="winrm access to gke kctf win build vm"
TIMEOUT=$((5*60))
SLEEP=20

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
  echo -n "[*] Trying to fetch windows credentials, remaining timeout: ${TIMEOUT}s"
  gcloud --quiet beta compute reset-windows-password "${VM_NAME}" > windows_creds 2>/dev/null && break
  if [[ ${TIMEOUT} -lt ${SLEEP} ]]; then
    echo ": failed, giving up"
    exit 1
  fi
  echo ": failed, sleeping ${SLEEP}s"
  sleep ${SLEEP}
  TIMEOUT=$((${TIMEOUT} - ${SLEEP}))
done
echo ': success'
