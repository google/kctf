#!/bin/bash

CHALLENGE="$1"

KEY="/pow-bypass/pow-bypass-key.pem"

SIG=$(echo -n "${CHALLENGE}" | openssl dgst -SHA256 -hex -sign "${KEY}" - | awk '{print $2}')

echo -n "b.${SIG}"
