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

# If there's a change in the CRD, we need to regenerate it and apply it to the cluster
#export GOROOT=$(go env GOROOT)
#operator-sdk generate k8s

IMAGE_URL="${REGISTRY}/${PROJECT}/kctf-operator"
echo "building image and pushing to ${IMAGE_URL}"

cd "${DIR}/kctf-operator"

operator-sdk build "${IMAGE_URL}"
OPERATOR_SHA=$(docker push "${IMAGE_URL}" | egrep -o 'sha256:[0-9a-f]+' | head -n1)
IMAGE_ID="${IMAGE_URL}@${OPERATOR_SHA}"
echo "pushed to ${IMAGE_ID}"
OPERATOR_YAML="${KCTF_CTF_DIR}/kctf/resources/operator.yaml"
sed -i "s#image: .*#image: ${IMAGE_ID}#" "${OPERATOR_YAML}"
"${KCTF_BIN}/kctf-cluster" start
