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
DIR="$( cd "$( dirname "$( readlink -f "${BASH_SOURCE[0]}")" )" >/dev/null && pwd )/../.."
source "${DIR}/scripts/lib/config.sh"

load_config

MIN_NODES="1"
MAX_NODES="2"
NUM_NODES="1"
MACHINE_TYPE="n2-standard-4"

EXISTING_CLUSTER=$(gcloud container clusters list --filter "name=${CLUSTER_NAME}" --format 'get(name)')

if [ -z "${EXISTING_CLUSTER}" ]; then
  gcloud container clusters create --release-channel=regular --enable-network-policy --enable-autoscaling --min-nodes ${MIN_NODES} --max-nodes ${MAX_NODES} --num-nodes ${NUM_NODES} --create-subnetwork name=kctf-${CLUSTER_NAME}-subnet --no-enable-master-authorized-networks --enable-ip-alias --enable-private-nodes --master-ipv4-cidr 172.16.0.32/28 --enable-autorepair --preemptible --machine-type ${MACHINE_TYPE} --workload-pool=${PROJECT}.svc.id.goog ${CLUSTER_NAME}
fi

EXISTING_ROUTER=$(gcloud compute routers list --filter "name=kctf-${CLUSTER_NAME}-nat-router" --format 'get(name)')
if [ -z "${EXISTING_ROUTER}" ]; then
  gcloud compute routers create "kctf-${CLUSTER_NAME}-nat-router" --network=default --region "${ZONE::-2}"
fi

EXISTING_NAT=$(gcloud compute routers nats list --router "kctf-${CLUSTER_NAME}-nat-router" --router-region "${ZONE::-2}" --format 'get(name)')
if [ -z "${EXISTING_NAT}" ]; then
  gcloud compute routers nats create "kctf-${CLUSTER_NAME}-nat-config" --router-region "${ZONE::-2}" --router kctf-${CLUSTER_NAME}-nat-router --nat-all-subnet-ip-ranges --auto-allocate-nat-external-ips
fi

get_cluster_creds

kubectl create namespace "kctf-system" --dry-run=client -oyaml | kubectl apply -f - >&2

# GCSFUSE

SUFFIX=$(echo "${PROJECT}-${CLUSTER_NAME}-${ZONE}" | sha1sum)
BUCKET_NAME="kctf-gcsfuse-${SUFFIX:0:16}"
GCS_GSA_NAME="${BUCKET_NAME}"
GCS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${GCS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)' || true)

if [ -z "${GCS_GSA_EMAIL}" ]; then
  gcloud iam service-accounts create "${GCS_GSA_NAME}" --description "kCTF GCSFUSE service account ${CLUSTER_NAME} ${ZONE}" --display-name "kCTF GCSFUSE ${CLUSTER_NAME} ${ZONE}"
  GCS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${GCS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)')
  while [ -z "${GCS_GSA_EMAIL}" ]; do
    sleep 1
    GCS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${GCS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)')
  done
fi

# Creating CRD, rbac and operator
kubectl apply -f "${DIR}/kctf-operator/deploy/crds/kctf.dev_challenges_crd.yaml"
kubectl apply -f "${DIR}/kctf-operator/deploy/rbac.yaml"
kubectl apply -f "${DIR}/kctf-operator/deploy/operator.yaml"

OPERATOR_IMAGE=$(yq read "${DIR}/kctf-operator/deploy/operator.yaml" 'spec.template.spec.containers[0].image')

# The operator needs to create some subresources, e.g. the gcsfuse service account
for i in {1..30}; do
  kctf-kubectl get pods --namespace kctf-system -o=jsonpath="{.items[*].status.containerStatuses[?(@.image==\"${OPERATOR_IMAGE}\")].ready}" | grep "true" && break
  if [ "$i" == "30" ]; then
    echo "Couldn't find a kctf-operator pod with status ready=true and image=\"${OPERATOR_IMAGE}\" after 5 minutes" >&2
    kctf-kubectl get pods --namespace kctf-system -o=yaml >&2
    exit 1
  fi
  sleep 10
done

GCS_KSA_NAME="gcsfuse-sa"

gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:${PROJECT}.svc.id.goog[kctf-system/${GCS_KSA_NAME}]" ${GCS_GSA_EMAIL}
kubectl annotate serviceaccount --namespace kctf-system ${GCS_KSA_NAME} iam.gke.io/gcp-service-account=${GCS_GSA_EMAIL} --overwrite

if ! gsutil du "gs://${BUCKET_NAME}/"; then
  gsutil mb -l eu "gs://${BUCKET_NAME}/"
fi

if gsutil uniformbucketlevelaccess get "gs://${BUCKET_NAME}" | grep -q "Enabled: True"; then
  gsutil iam ch "serviceAccount:${GCS_GSA_EMAIL}:roles/storage.legacyBucketOwner" "gs://${BUCKET_NAME}"
  gsutil iam ch "serviceAccount:${GCS_GSA_EMAIL}:roles/storage.legacyObjectOwner" "gs://${BUCKET_NAME}"
else
  gsutil acl ch -u "${GCS_GSA_EMAIL}:O" "gs://${BUCKET_NAME}"
fi

kubectl create configmap gcsfuse-config --from-literal=gcs_bucket="${BUCKET_NAME}" --namespace kctf-system --dry-run=client -o yaml | kubectl apply -f -

kubectl patch ServiceAccount default --patch "automountServiceAccountToken: false"

# Cloud DNS

if [ ! -z "${DOMAIN_NAME}" ]
then
  DNS_ZONE=$(gcloud dns managed-zones list --filter "dns_name:${DOMAIN_NAME}" --format 'get(name)')
  if [ -z "${DNS_ZONE}" ]; then
    gcloud dns managed-zones create "${CLUSTER_NAME}-dns-zone" --description "DNS Zone for ${DOMAIN_NAME}" --dns-name="${DOMAIN_NAME}."
  fi

  DNS_GSA_NAME="kctf-cloud-dns"
  DNS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${DNS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)' || true)

  if [ -z "${DNS_GSA_EMAIL}" ]; then
    gcloud iam service-accounts create "${DNS_GSA_NAME}" --description "kCTF Cloud DNS service account ${CLUSTER_NAME} ${ZONE}" --display-name "kCTF Cloud DNS ${CLUSTER_NAME} ${ZONE}"
    DNS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${DNS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)')
    while [ -z "${DNS_GSA_EMAIL}" ]; do
      sleep 1
      DNS_GSA_EMAIL=$(gcloud iam service-accounts list --filter "email=${DNS_GSA_NAME}@${PROJECT}.iam.gserviceaccount.com" --format 'get(email)')
    done
  fi

  DNS_KSA_NAME="external-dns-sa"

  gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:${PROJECT}.svc.id.goog[kctf-system/${DNS_KSA_NAME}]" ${DNS_GSA_EMAIL}
  kubectl annotate serviceaccount --namespace kctf-system ${DNS_KSA_NAME} iam.gke.io/gcp-service-account=${DNS_GSA_EMAIL} --overwrite

  gcloud projects add-iam-policy-binding ${PROJECT} --member=serviceAccount:${DNS_GSA_EMAIL} --role=roles/dns.admin

  kubectl create configmap --namespace kctf-system external-dns --from-literal=DOMAIN_NAME=${DOMAIN_NAME} --dry-run=client -o yaml | kubectl apply -f -
fi
