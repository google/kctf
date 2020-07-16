#!/bin/bash

set -Eeuo pipefail
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )"

if [ -z "${PROJECT}" ]; then echo 'missing PROJECT'; exit 1; fi
if [ -z "${CHALLENGE_NAME}" ]; then echo 'missing CHALLENGE_NAME'; exit 1; fi
if [ -z "${REGISTRY}" ]; then echo 'missing REGISTRY'; exit 1; fi
BUCKET="gs://${PROJECT}-remote-build"
#TODO reusing the gs path like this means we can't run in parallel
GS_PATH="${BUCKET}/remote_build/${CHALLENGE_NAME}"
IMAGE_TAG="${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}"

CONTAINER_DIR="$1"

gsutil ls "${BUCKET}" >/dev/null 2>/dev/null || gsutil mb "${BUCKET}" >&2
gsutil rm -R "${GS_PATH}" >/dev/null 2>/dev/null || true
gsutil cp -R "${CONTAINER_DIR}/image" "${GS_PATH}/" >&2

VENV="${DIR}/venv"

if [ ! -d "${VENV}" ]; then
  virtualenv "${VENV}" >&2
  PS1=""
  source "${VENV}/bin/activate" >&2
  pip install pywinrm >&2
else
  PS1=""
  source "${VENV}/bin/activate" >&2
fi

"${DIR}/build_remote.py" "${GS_PATH}" "${IMAGE_TAG}" "${DIR}/windows_creds"
