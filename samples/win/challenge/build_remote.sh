#!/bin/bash

set -Eeuo pipefail

if [ -z "${PROJECT}" ]; then echo 'missing PROJECT'; exit 1; fi
if [ -z "${CHALLENGE_NAME}" ]; then echo 'missing CHALLENGE_NAME'; exit 1; fi
if [ -z "${REGISTRY}" ]; then echo 'missing REGISTRY'; exit 1; fi
BUCKET="gs://${PROJECT}-remote-build"
GS_PATH="${BUCKET}/remote_build/${CHALLENGE_NAME}"
IMAGE_TAG="${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}"

gsutil ls "${BUCKET}" >/dev/null 2>/dev/null || gsutil mb "${BUCKET}" >&2
gsutil rm -R "${GS_PATH}" >/dev/null 2>/dev/null || true
gsutil cp -R image/* "${GS_PATH}/" >&2

if [ ! -d venv ]; then
  virtualenv venv >&2
  PS1=""
  source venv/bin/activate >&2
  pip install pywinrm >&2
else
  PS1=""
  source venv/bin/activate >&2
fi

./build_remote.py "${GS_PATH}" "${IMAGE_TAG}"
