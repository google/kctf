#!/bin/bash

if [ "$#" -lt 3 ]; then
  echo "Illegal number of parameters"
  echo "Usage: docker.sh LOCAL_IMAGE REMOTE_IMAGE VERSION [PUSH]"
  echo
  echo "For example: docker.sh pwntools gcr.io/kctf-docker/pwntools master true"
  echo "  That command will tag and push to GCR the local image pwntools."
fi

IMAGE=$1
REPO=$2
REF=$3
PUSH=${4-false}

# Strip git ref prefix from version
VERSION=$(echo "$REF" | sed -e 's,.*/\(.*\),\1,')

# Strip "v" prefix from tag name
[[ "$REF" == "refs/tags/"* ]] && VERSION=$(echo $VERSION | sed -e 's/^v//')

# Use Docker `latest` tag convention
[ "$VERSION" == "master" ] && VERSION=latest

docker tag $IMAGE $REPO:$VERSION

[ "$PUSH" == "true"] && docker push $REPO:$VERSION
