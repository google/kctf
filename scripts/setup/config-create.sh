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

mkdir -p "${CONFIG_DIR}"

# Default Configuration

CHAL_DIR=""
REGISTRY="eu.gcr.io"
PROJECT=""
ZONE="europe-west4-b"
CLUSTER_NAME="kctf-cluster"
DOMAIN_NAME=""
START_CLUSTER="0"

if [ -d ${CONFIG_DIR}/challenges ]; then
    load_chal_dir
fi

function usage {
    echo "$(basename $0) [args]" >&2
    echo -e "\t--chal-dir\tdirectory in which challenges are stored (default: \"${CHAL_DIR}\")" >&2
    echo -e "\t--project\tGoogle Cloud Platform project name" >&2
    echo -e "\t--zone\t\tGCP Zone (default: europe-west4-b)" >&2
    echo -e "\t\t\tFor a list of zones run: gcloud compute machine-types list --filter=\"name=( n2-standard-4 )\" --format 'value(zone)'" >&2
    echo -e "\t--registry\t\tContainer Registry (default: eu.gcr.io)" >&2
    echo -e "\t\t\tPossible values are us.gcr.io, asia.gcr.io, and eu.gcr.io" >&2
    echo -e "\t--cluster-name\tName of the kubernetes cluster (default: kctf-cluster)" >&2
    echo -e "\t--domain-name\tOptional domain name to host challenges under" >&2
    echo -e "\t--start-cluster\tStart the cluster if it's not running yet" >&2
    exit 1
}

while :; do
    if [[ -z "${1:-}" ]]; then
        break
    fi
    if [[ "$1" != "--start-cluster" ]] && [[ -z "${2:-}" ]]; then
        echo "Missing argument after \"$1\"." >&2
        usage
    fi
    case $1 in
        --chal-dir)
            CHAL_DIR=$2
            if [[ $CHAL_DIR == ~\/* ]]; then
                CHAL_DIR="${HOME}/${CHAL_DIR:2}"
            fi
            ;;
        --project)
            PROJECT=$2
            ;;
        --zone)
            ZONE=$2
            ;;
        --registry)
            REGISTRY=$2
            ;;
        --cluster-name)
            CLUSTER_NAME=$2
            ;;
        --domain-name)
            DOMAIN_NAME="$2"
            ;;
        --start-cluster)
            START_CLUSTER="1"
            ;;
        *)
            echo "Unrecognized argument \"$1\"." >&2
            usage
            ;;
    esac
    if [[ "$1" != "--start-cluster" ]]; then
        shift
    fi
    shift
done

if [[ -z "$PROJECT" ]]; then
    echo "Missing required argument \"--project\"." >&2
    usage
fi

"${DIR}/scripts/setup/challenge-directory.sh" "${CHAL_DIR}"
load_chal_dir

generate_config_dir
CLUSTER_DIR="${ret}"
mkdir -p "${CLUSTER_DIR}"

generate_config_path
CLUSTER_CONFIG="${ret}"
cat > "${CLUSTER_CONFIG}" << EOF
PROJECT=${PROJECT}
ZONE=${ZONE}
REGISTRY=${REGISTRY}
CLUSTER_NAME=${CLUSTER_NAME}
DOMAIN_NAME=${DOMAIN_NAME}
EOF

ln -fs "${CLUSTER_CONFIG}" "${CONFIG_FILE}"

# checks if gcloud is installed, if not, it creates only locally
if command -v gcloud >/dev/null 2>&1; then
    create_gcloud_config
else
    echo "Configuration created only locally. Gcloud not installed." >&2
fi
    
create_gcloud_config

# there might be an existing cluster
# if it already exists, we try to update it
# otherwise, start it if requsted
if [[ "${START_CLUSTER}" == "1" ]] || get_cluster_creds 2>/dev/null; then
    "${DIR}/scripts/cluster/start.sh"
fi
