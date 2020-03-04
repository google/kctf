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

if [ $# != 1 ]; then
    echo "usage: $0 /path/to/challenge/directory"
    exit 1
fi

CHAL_DIR=$1

if [ ! -d "${CHAL_DIR}" ]; then
    if [ -e "${CHAL_DIR}" ]; then
        echo "error: ${CHAL_DIR} already exists and is not a directory"
        exit 1
    fi
    echo "creating ${CHAL_DIR}"
fi

CHAL_DIR=$(realpath -L ${CHAL_DIR})

# copy the base files to the chal dir
if [ ! -d "${CHAL_DIR}/kctf-conf" ]; then
    mkdir -p "${CHAL_DIR}/kctf-conf"
    umask o+rx
    cp -p -R "${DIR}/base" "${CHAL_DIR}/kctf-conf/"
fi

mkdir -p "${CONFIG_DIR}"
ln -nfs "${CHAL_DIR}" "${CONFIG_DIR}/challenges"
if [ -f "${CONFIG_DIR}/cluster.conf" ]; then
    source "${CONFIG_DIR}/cluster.conf"

    CONFIG_PATH=$(readlink "${CONFIG_DIR}/cluster.conf")
    generate_config_path
    if [ "${ret}" != "${CONFIG_PATH}" ]; then
        echo "unsetting cluster.conf from different challenge directory: ${CONFIG_PATH}"
        rm "${CONFIG_DIR}/cluster.conf"
    fi
else
    # remove dangling link
    if [ -L "${CONFIG_DIR}/cluster.conf" ]; then
        rm "${CONFIG_DIR}/cluster.conf"
    fi
fi

DOCKER_LOGS=$(mktemp -d)
DOCKER_PIDS=""

if [ -z "$(docker images -f reference=kctf-nsjail-bin -q)" ]; then
  docker build -t kctf-nsjail-bin "${DIR}/config/docker" -f "${DIR}/config/docker/nsjail.Dockerfile" \
    >${DOCKER_LOGS}/nsjail.STDOUT 2>${DOCKER_LOGS}/nsjail.STDERR &
  DOCKER_PIDS="$DOCKER_PIDS $!"
fi

if [ -z "$(docker images -f reference=kctf-nsjail-chroot -q)" ]; then
  docker build -t kctf-nsjail-chroot "${DIR}/config/docker" -f "${DIR}/config/docker/chroot.Dockerfile" \
    >${DOCKER_LOGS}/chroot.STDOUT 2>${DOCKER_LOGS}/chroot.STDERR &
  DOCKER_PIDS="$DOCKER_PIDS $!"
fi

if [ -z "$(docker images -f reference=kctf-pwntools -q)" ]; then
  echo '==='
  docker build -t kctf-pwntools "${DIR}/config/docker" -f "${DIR}/config/docker/pwntools.Dockerfile" \
    >${DOCKER_LOGS}/pwntools.STDOUT 2>${DOCKER_LOGS}/pwntools.STDERR &
  DOCKER_PIDS="$DOCKER_PIDS $!"
fi

if [ "${DOCKER_BACKGROUND-false}" == "true" ]; then
  echo "== NOTICE: Building docker images in the background =="
  echo "   pids: ${DOCKER_PIDS:1} tmp: ${DOCKER_LOGS}"
else
  { cd ${DOCKER_LOGS} && tail -f * & }
  trap 'kill $(jobs -p)' EXIT
  for pid in ${DOCKER_PIDS}; do
    wait "$pid"
  done;
  rm -rf ${DOCKER_LOGS}
fi
