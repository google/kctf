#!/bin/bash

if [ -z "${DOMAIN}" ]; then
  echo "Make sure the DOMAIN environment variable points to the domain."
  exit 1
fi

if [ -z "${SECRET}" ]; then
  echo "Make sure the SECRET environment variable points to the k8s secret."
  exit 1
fi

if [ -z "${PROD}" ]; then
  echo "Making a TEST certificate because PROD environment variable is NOT set."
  TEST="--test-cert"
else
  echo "Making a REAL certificate because PROD environment variable is set."
  TEST=""
fi

if [ -z "${EMAIL}" ]; then
  echo "Registering certificate unsafely without email. Pass an EMAIL to register an account with an email address."
  EMAIL_FLAG="--register-unsafely-without-email"
else
  EMAIL_FLAG="-m ${EMAIL}"
fi

function request_certificate() {
  certbot certonly ${TEST} --non-interactive --agree-tos ${EMAIL_FLAG} --dns-google -d '*.'"${DOMAIN}" --dns-google-propagation-seconds 120
}

function update_tls_secret() {
  ./kubectl create secret tls "${SECRET}" --cert /etc/letsencrypt/live/"${DOMAIN}"/fullchain.pem --key /etc/letsencrypt/live/"${DOMAIN}"/privkey.pem --namespace kctf-system --dry-run=client --save-config -o yaml | ./kubectl apply -f -
}

function check_tls_validity() {
  ./kubectl get secret "${SECRET}" --namespace kctf-system -o 'jsonpath={.data}' | jq -r '.["tls.crt"]' | base64 -d | openssl x509 -checkend 2592000 -noout -in -
}

while true; do
  echo "Waiting 2 minutes to avoid hitting rate limits"
  sleep 2m
  if check_tls_validity; then
    echo "Certificate is valid for at least 30 days"
  else
    request_certificate && update_tls_secret && echo "TLS cert updated"
  fi
done
