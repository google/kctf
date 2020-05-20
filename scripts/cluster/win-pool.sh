#!/bin/bash
# Copyright 2020 Google LLC
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     https://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -Eeuo pipefail
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )/../.."
source "${DIR}/scripts/lib/config.sh"

load_config

MACHINE_TYPE="n2-standard-4"
MIN_NODES="1"
MAX_NODES="2"
NUM_NODES="1"
IMAGE_TYPE="WINDOWS_SAC"
POOL_NAME="win"
#IMAGE_TYPE="WINDOWS_LTSC"

function usage {
    echo "$(basename $0) [args]" >&2
    echo -e "\t--machine-type\tMachine type to use for the windows pool (default: \"${MACHINE_TYPE}\")" >&2
    echo -e "\t--min-nodes\tMinimum number of nodes" >&2
    echo -e "\t--max-nodes\tMaximum number of nodes" >&2
    echo -e "\t--num-nodes\tInitial number of nodes" >&2
    echo -e "\t--image-type\tWindows image type (e.g. WINDOWS_SAC or WINDOWS_LTSC)" >&2
    echo -e "\t--pool-name\tName of the node pool (default: \"${POOL_NAME}\")" >&2
    exit 1
}

while :; do
    if [[ -z "${1:-}" ]]; then
        break
    fi
    if [[ -z "${2:-}" ]]; then
        echo "Missing argument after \"$1\"." >&2
        usage
    fi
    case $1 in
        --machine-type)
            MACHINE_TYPE=$2
            ;;
        --min-nodes)
            MIN_NODES=$2
            ;;
        --max-nodes)
            MAX_NODES=$2
            ;;
        --num-nodes)
            NUM_NODES=$2
            ;;
        --image-type)
            IMAGE_TYPE=$2
            ;;
        --pool-name)
            POOL_NAME=$2
            ;;
        *)
            echo "Unrecognized argument \"$1\"." >&2
            usage
            ;;
    esac
    shift
done

gcloud container node-pools create "${POOL_NAME}" \
  --cluster="${CLUSTER_NAME}" \
  --machine-type="${MACHINE_TYPE}" \
  --enable-autorepair \
  --enable-autoupgrade \
  --num-nodes="${NUM_NODES}" \
  --enable-autoscaling \
  --min-nodes="${MIN_NODES}" \
  --max-nodes="${MAX_NODES}" \
  --no-enable-autoupgrade \
  --image-type="${IMAGE_TYPE}"
