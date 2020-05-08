#!/bin/bash

set -Eeuo pipefail

# if you change this, also update start-build-vm.sh
VM_NAME="win-build-vm"

gcloud beta compute instances delete "${VM_NAME}"
