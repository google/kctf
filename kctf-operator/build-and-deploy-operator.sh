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
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )/.."

if [[ -z "${PROJECT}" ]]; then
  echo "you need to load a cluster config first (source kctf/activate)" >&2
  exit 1
fi

IMAGE_BASE="${REGISTRY}/${PROJECT}"
echo "building images and pushing to ${IMAGE_BASE}"

pushd "${DIR}/kctf-operator"

set -x

GCSFUSE_IMAGE_URL="${IMAGE_BASE}/gcsfuse"
CERTBOT_IMAGE_URL="${IMAGE_BASE}/certbot"

GCSFUSE_IMAGE_ID=$(docker build -t "${GCSFUSE_IMAGE_URL}" -q "${DIR}/docker-images/gcsfuse")
CERTBOT_IMAGE_ID=$(docker build -t "${CERTBOT_IMAGE_URL}" -q "${DIR}/docker-images/certbot")

docker push "${GCSFUSE_IMAGE_URL}"
docker push "${CERTBOT_IMAGE_URL}"

GCSFUSE_IMAGE_SHA=$(docker inspect -f '{{index .RepoDigests 0}}' "${GCSFUSE_IMAGE_ID}")
CERTBOT_IMAGE_SHA=$(docker inspect -f '{{index .RepoDigests 0}}' "${CERTBOT_IMAGE_ID}")

sed -i 's%const DOCKER_GCSFUSE_IMAGE = .*%const DOCKER_GCSFUSE_IMAGE = "'${GCSFUSE_IMAGE_SHA}'"%' resources/constants.go
sed -i 's%const DOCKER_CERTBOT_IMAGE = .*%const DOCKER_CERTBOT_IMAGE = "'${CERTBOT_IMAGE_SHA}'"%' resources/constants.go

set +x

IMAGE_URL="${IMAGE_BASE}/kctf-operator:dev"
make manifests docker-build operator.yaml bundle IMG="${IMAGE_URL}"
OPERATOR_SHA=$(docker push "${IMAGE_URL}" | egrep -o 'sha256:[0-9a-f]+' | head -n1)
IMAGE_ID="${IMAGE_BASE}/kctf-operator@${OPERATOR_SHA}"

echo "pushed to ${IMAGE_ID}"

OPERATOR_YAML="${KCTF_CTF_DIR}/kctf/resources/operator.yaml"
sed -i "s#image: ${IMAGE_URL}#image: ${IMAGE_ID}#" "operator.yaml"
cat "operator.yaml" > "${OPERATOR_YAML}"

cp -f bundle/manifests/*.yaml "${KCTF_CTF_DIR}/kctf/resources/"

popd

"${KCTF_BIN}/kctf-cluster" start
