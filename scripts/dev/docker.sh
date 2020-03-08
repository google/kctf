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

if [ "$#" -lt 3 ]; then
  echo "Illegal number of parameters"
  echo "Usage: docker.sh LOCAL_IMAGE REMOTE_IMAGE VERSION [PUSH]"
  echo
  echo "For example: docker.sh pwntools gcr.io/kctf-docker/pwntools master true"
  echo "  That command will tag and push to GCR the local image pwntools."
  echo "  If you ommit the last parameter (true), then the challenge won't be pushed."
  exit 1
fi

IMAGE=$1
REPO=$2
REF=$3
PUSH=${4-false}

VERSION=$(echo "$REF" | sed -e 's,.*/\(.*\),\1,')
[[ "$REF" == "refs/tags/"* ]] && VERSION=$(echo $VERSION | sed -e 's/^v//')
[ "$VERSION" == "master" ] && VERSION=latest

docker tag $IMAGE $REPO:$VERSION
[ "$PUSH" == "true"] && docker push $REPO:$VERSION
