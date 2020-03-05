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
PROJECT=""
ZONE="europe-west4-b"
CLUSTER_NAME="kctf-cluster"
DOMAIN_NAME=""

config=""
read_config() {
    read -e -p "  $2: " -i "${!1}" "$1"
    line="${1}=${!1}"
    config="${config}"$'\n'"${line}"
}

if [ -d ${CONFIG_DIR}/challenges ]; then
    load_chal_dir
    echo
    echo "Configuring cluster for challenge directory ${CHAL_DIR}"
else
    echo
    read -e -p "In which directory will challenges be stored?: " "CHAL_DIR"

    if [[ $CHAL_DIR == ~\/* ]]; then
        CHAL_DIR="${HOME}/${CHAL_DIR:2}"
    fi

    "${DIR}/scripts/setup/challenge-directory.sh" "${CHAL_DIR}"

    load_chal_dir
fi

echo
echo "= CLUSTER CONFIGURATION ="
echo
if test -f "${CONFIG_FILE}"; then
    echo
    echo " Reusing the last config file used ($(readlink -f "${CONFIG_FILE}"))."
    echo
    . "${CONFIG_FILE}"
else
    echo
    echo " Creating a new config file from scratch."
    echo
fi
echo
echo "== PROJECT NAME =="
echo
echo " Important: Make sure to update this field to your own project, or nothing will work"
echo
read_config PROJECT "Google Cloud Platform project name"
gcloud config set project "${PROJECT}"
echo
echo "= OPTIONAL CONFIGURATION ="
echo
echo " Note: You can leave everything below here as defaults."
echo
echo "== ZONE =="
echo
echo " Used for cluster configuration"
echo "  The zone defines the geographic location of the cluster."
echo
echo -n "  Available zones: (Loading...)"
gcloud compute machine-types list --filter="name=( n2-standard-4 )" --format 'value(zone)' | xargs echo -e "\r  Available zones: "
echo
read_config ZONE "GCP Zone"
echo
echo "== CLUSTER NAME =="
echo
echo "  If you are reusing the same project for multiple CTFs, make sure this name is unique."
echo
read_config CLUSTER_NAME "Name of the cluster"
echo
echo "== DOMAIN NAME =="
echo
echo "  If you want to configure a domain name for the challenge, provide it below, otherwise leave it empty."
echo
read_config DOMAIN_NAME "Domain name (eg, k8s.ctfcompetititon.com)"
echo

CLUSTER_DIR="${CHAL_DIR}/kctf-conf/${PROJECT}_${ZONE}_${CLUSTER_NAME}"
CLUSTER_CONFIG="${CLUSTER_DIR}/cluster.conf"

echo "= SUMMARY ="
echo " This is the configuration for your cluster, please review it to make sure it is correct. It will be written to ${CLUSTER_CONFIG}"
echo "$config"
echo
echo " If you wish to change anything, just run this command again."

mkdir -p "${CLUSTER_DIR}"
echo "${config}" > "${CLUSTER_CONFIG}"

ln -fs "${CLUSTER_CONFIG}" "${CONFIG_FILE}"

create_gcloud_config

# there might be an existing cluster
# if it already exists, we try to update it
# otherwise, ask if we should start it
if get_cluster_creds 2>/dev/null; then
    "${DIR}/scripts/cluster/start.sh"
else
    echo
    read -p "Start the cluster now (y/N)? " SHOULD_START
    echo
    if [[ ${SHOULD_START} =~ ^[Yy]$ ]]; then
        "${DIR}/scripts/cluster/start.sh"
    fi
fi
