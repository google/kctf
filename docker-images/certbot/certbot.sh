#!/bin/bash

TEST="--test-cert"

if [ -z "${DOMAIN}" ]; then
  echo Make sure the DOMAIN environment variable points to the domain.
  exit 1
fi

if [ -z "${EMAIL}" ]; then
  echo Make sure the EMAIL environment variable points to the email.
  exit 1
fi

if [ -z "${SECRET}" ]; then
  echo Make sure the SECRET environment variable points to the k8s secret.
  exit 1
fi

if [ -z "${PROD}" ]; then
  echo Making a test certificate because PROD environment variable is not set.
else
  echo Making a valid certificate because PROD environment variable is set.
  TEST=""
fi

while true; do
  echo Waiting 2 minutes to avoid hitting rate limits
  sleep 2m
  if ./kubectl get secret "${SECRET}" -o 'jsonpath={.data}' | jq -r '.["tls.crt"]' | base64 -d | openssl x509 -checkend 2592000 -noout -in -; then
    echo Certificate is valid for at least 30 days
    sleep 15d
  else
    certbot certonly "${TEST}" --non-interactive --agree-tos -m "${EMAIL}" --dns-google -d '*.'"${DOMAIN}" --dns-google-propagation-seconds 120 && \
    (./kubectl create secret tls "${SECRET}" --cert /etc/letsencrypt/live/"${DOMAIN}"/fullchain.pem --key /etc/letsencrypt/live/"${DOMAIN}"/privkey.pem --dry-run=client --save-config -o yaml | ./kubectl apply -f -) && \
    echo Created and saved certificate
  fi
done
