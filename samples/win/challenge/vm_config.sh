VM_TAG="win-build-vm"
MACHINE_TYPE="n1-standard-4"
IMAGE="windows-server-1909-dc-core-for-containers-v20200414"
VM_NAME="win-build-vm"
FIREWALL_RULE="gke-kctf-win-build-vm"
FIREWALL_DESCRIPTION="winrm access to gke kctf win build vm"
CREDENTIAL_FETCH_TIMEOUT=$((5*60))
CREDENTIAL_FETCH_SLEEP=20
