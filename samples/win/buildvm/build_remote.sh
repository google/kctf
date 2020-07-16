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
gsutil cp -R "${CONTAINER_DIR}/build.ps1" "${GS_PATH}/" >&2

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

export CREDS_PATH="${DIR}/windows_creds"

"${DIR}/run_on_win.py" rmdir 'c:\build' /s /q >/dev/null 2>/dev/null || true
"${DIR}/run_on_win.py" mkdir 'c:\build' >&2
"${DIR}/run_on_win.py" gsutil -m cp -r "${GS_PATH}/*" 'c:\build' >&2
"${DIR}/run_on_win.py" gcloud auth configure-docker --quiet >&2
TMP_TAG="${IMAGE_TAG}:tmp"
"${DIR}/run_on_win.py" powershell.exe -file 'C:\build\build.ps1' >&2
"${DIR}/run_on_win.py" docker build -t "${TMP_TAG}" 'c:\build\image' >&2
DIGEST=$("${DIR}/run_on_win.py" docker image ls -q "${TMP_TAG}")
FULL_TAG="${IMAGE_TAG}:${DIGEST}"
"${DIR}/run_on_win.py" docker tag "$TMP_TAG" "${FULL_TAG}" >&2
"${DIR}/run_on_win.py" docker push "${FULL_TAG}" >&2
echo -n "${FULL_TAG}"
