#
#
#
# THIS MAKEFILE IS USED CTF-WIDE. IF YOU MAKE CHANGES IT WILL APPLY TO ALL CHALLENGES.
#
#
#
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

.PHONY: start stop docker ip status logs healthcheck-logs ssh healthcheck-ssh port-forward test-docker healthcheck-test-docker test-kind test-d4w .test-local clean .deploy .network .https .cluster-config .deployment .FORCE

SHELL := bash
.ONESHELL:
.SHELLFLAGS = -e -c

PUSH_TARGET="REMOTE"

REGISTRY:=CONFIGMISSING
PROJECT:=CONFIGMISSING
CLUSTER_NAME:=CONFIGMISSING
ZONE:=CONFIGMISSING
# sets DEPLOY, PUBLIC, HEALTHCHECK, HTTPS
include config/chal.conf
-include $(HOME)/.config/kctf/cluster.conf
KUBECONFIG=$(HOME)/.config/kctf/kube.conf
export KUBECONFIG

CHALLENGE_NAME:=$(shell basename ${CURDIR})
CLUSTER_GEN=.gen/${PROJECT}_${ZONE}_${CLUSTER_NAME}

ifeq ($(HEALTHCHECK), true)
MAYBE_REMOTE_HEALTHCHECK_IMAGE=${CLUSTER_GEN}/remote-healthcheck-image
MAYBE_HEALTHCHECK=${CLUSTER_GEN}/healthcheck-image-pushed ${CLUSTER_GEN}/remote-healthcheck-image
else
MAYBE_REMOTE_HEALTHCHECK_IMAGE=
MAYBE_HEALTHCHECK=
endif

docker: .gen/challenge-image

start:
ifeq ($(DEPLOY), true)
	  $(MAKE) .deploy
else
	  echo "skipping deployment: DEPLOY=\"$${DEPLOY}\""
endif

stop: .cluster-config
	kubectl delete namespace ${CHALLENGE_NAME}

ip: .cluster-config
	@LB_IP=""
	@while [ -z "$${LB_IP}" ]; do
	@  LB_IP=$$(kubectl get "service/chal" -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
	@  sleep 3
	@done
	@echo "$${LB_IP}"

status: .cluster-config
	@echo "= INSTANCES / PODs ="
	@echo
	@echo "Challenge execution status"
	@echo "This shows you how many instances of the challenges are running."
	@echo
	@kubectl get pods -l "app=${CHALLENGE_NAME}" -o wide
	@echo
	@echo
	@echo "= DEPLOYMENTS ="
	@echo
	@echo "Challenge deployment status"
	@echo "This shows you if the challenge was deployed to the cluster."
	@echo
	@kubectl get deployments -l "app=${CHALLENGE_NAME}" -o wide
	@echo
	@echo "= EXTERNAL SERVICES ="
	@echo
	@echo "Challenge external status"
	@echo "This shows you if the challenge is exposed externally."
	@echo
	@kubectl get services -l "app=${CHALLENGE_NAME}" -o wide
	@kubectl get ingress -l "app=${CHALLENGE_NAME}" -o wide
	@echo

logs: .cluster-config
	kubectl logs -l "app=${CHALLENGE_NAME}" -c challenge

healthcheck-logs: .cluster-config
	kubectl logs -l "app=${CHALLENGE_NAME}" -c healthcheck

ssh: .cluster-config
	kubectl exec deployment/chal -c challenge -it -- /bin/bash

healthcheck-ssh: .cluster-config
	kubectl exec deployment/chal -c healthcheck -it -- /bin/bash

port-forward: .cluster-config
	kubectl port-forward deployment/chal :1337 &

test-docker: .gen/docker-id
	docker ps -f "id=$$(cat .gen/docker-id)"

healthcheck-test-docker: .gen/healthcheck-docker-id
	docker ps -f "id=$$(Icat .gen/healthcheck-docker-id)"

.gen/healthcheck-docker-id: .gen/docker-name .gen/healthcheck-image
	IMAGE_ID="$$(cat .gen/healthcheck-image)"
	test -s .gen/healthcheck-docker-id && docker kill $$(cat .gen/healthcheck-docker-id) || true
	docker run -d -it --network "container:$$(cat .gen/docker-name)" \
		"$${IMAGE_ID}" > $@

.gen/docker-name: .gen/docker-id
	docker ps --all -f id=$$(cat .gen/docker-id) --format {{.Names}} > $@

.gen/docker-id: PUSH_TARGET="LOCAL_DOCKER"
.gen/docker-id: .gen/challenge-image
	test -s .gen/docker-id && docker kill $$(cat .gen/docker-id) || true
	IMAGE_ID="$$(cat .gen/challenge-image)"
	docker run -d -it -p 1337 --privileged "$${IMAGE_ID}" > $@

test-kind: LOCAL_KUBECONFIG=".gen/kind-kubeconfig"
test-kind: PUSH_TARGET="KIND"
test-kind: | .gen/kind-kubeconfig .test-local

test-d4w: LOCAL_KUBECONFIG=".gen/d4w-kubeconfig"
test-d4w: PUSH_TARGET="LOCAL_DOCKER"
test-d4w: | .gen/d4w-kubeconfig .test-local

.test-local: PROJECT="kctf-testing"
.test-local:
	kubectl config rename-context "$(shell kubectl config current-context --kubeconfig="${LOCAL_KUBECONFIG}")" "kctf_${PROJECT}_${ZONE}_${CLUSTER_NAME}" --kubeconfig="${LOCAL_KUBECONFIG}" || true
	$(MAKE) .deploy PUSH_TARGET=${PUSH_TARGET} PROJECT=${PROJECT} ZONE=${ZONE} CLUSTER_NAME=${CLUSTER_NAME} KUBECONFIG="${LOCAL_KUBECONFIG}"
	kubectl patch deployment chal  --type json -p='[{"op": "remove", "path": "/spec/template/metadata/annotations/container.apparmor.security.beta.kubernetes.io~1challenge"}]' --kubeconfig=${LOCAL_KUBECONFIG}

clean:
	rm -R .gen/* || true
	rm -R challenge/.gen/* || true
	rm -R healthcheck/.gen/* || true

.deploy: .deployment .cluster-config .network

define CERTIFICATE_TEMPLATE
apiVersion: "networking.gke.io/v1beta1"
kind: "ManagedCertificate"
metadata:
  name: "kctf-certificate"
spec:
  domains:
endef

define INGRESS_TEMPLATE
apiVersion: "networking.k8s.io/v1beta1"
kind: "Ingress"
metadata:
  name: "https"
  labels:
    app: "$(CHALLENGE_NAME)"
  annotations:
    networking.gke.io/managed-certificates: "kctf-certificate"
spec:
  backend:
    serviceName: "chal"
    servicePort: 1
endef

export CERTIFICATE_TEMPLATE
export INGRESS_TEMPLATE

.https: .cluster-config
ifeq ($(HTTPS), true)
ifneq ($(DNS), true)
	$(error 'HTTPS=true but DNS=false. Need DNS to get certificate.')
endif
ifeq ($(DOMAIN_NAME),)
        $(error 'HTTPS=true but DOMAIN_NAME is not set.')
endif
	(echo "$${CERTIFICATE_TEMPLATE}"; echo '    -' "$(CHALLENGE_NAME).$(DOMAIN_NAME)") | kubectl apply -f -
	echo "$${INGRESS_TEMPLATE}" | kubectl apply -f -
	kubectl annotate "ingress/https" "external-dns.alpha.kubernetes.io/hostname=${CHALLENGE_NAME}-http.${DOMAIN_NAME}" --overwrite
else
	@(kubectl get "ingress/https" && kubectl delete "ingress/https" || true) 2>/dev/null
endif

.network: .gen/k8s .https | .cluster-config
ifeq ($(PUBLIC), true)
	kubectl apply -k config/advanced/network
ifeq ($(DNS), true)
ifeq ($(DOMAIN_NAME),)
	$(error 'DNS=true but DOMAIN_NAME is not set.')
endif
	kubectl get "service/chal" >/dev/null 2>&1 && kubectl annotate "service/chal" "external-dns.alpha.kubernetes.io/hostname=${CHALLENGE_NAME}-tcp.${DOMAIN_NAME}" --overwrite
endif
else
	kubectl get "service/chal" >/dev/null 2>&1 && kubectl delete "service/chal" || true
endif

.deployment: .gen/k8s ${CLUSTER_GEN}/image-pushed ${CLUSTER_GEN}/remote-image ${MAYBE_HEALTHCHECK} | .cluster-config
	kubectl apply -k config/advanced/deployment
	kubectl create configmap pow --from-file=pow.conf=config/pow.conf --dry-run=client -o yaml | kubectl apply -f -
	kubectl create secret generic pow-bypass --from-file="../kctf-conf/secrets/pow-bypass-key.pem" --dry-run=client -o yaml | kubectl apply -f -
	kubectl create secret generic pow-bypass-pub --from-file="../kctf-conf/secrets/pow-bypass-key-pub.pem" --dry-run=client -o yaml | kubectl apply -f -
# update the challenge container if the image changed
	PUSHED_IMAGE="$$(cat ${CLUSTER_GEN}/image-tagged)"
	CHAL_IMAGE="$$(cat ${CLUSTER_GEN}/remote-image)"
	if [ "$${CHAL_IMAGE}" != "$${PUSHED_IMAGE}" ]; then
	  kubectl set image "deployment/chal" "challenge=$${PUSHED_IMAGE}"
	fi
ifeq ($(HEALTHCHECK), true)
# update the healthcheck container if the image changed
	PUSHED_HEALTHCHECK_IMAGE="$$(cat ${CLUSTER_GEN}/healthcheck-image-tagged)"
	HEALTHCHECK_IMAGE="$$(cat ${CLUSTER_GEN}/remote-healthcheck-image)"
	if [ "$${HEALTHCHECK_IMAGE}" != "$${PUSHED_HEALTHCHECK_IMAGE}" ]; then
	  kubectl set image "deployment/chal" "healthcheck=$${PUSHED_HEALTHCHECK_IMAGE}"
	fi
endif

.gen/challenge-image: challenge/.gen/docker-image
	NEW_ID=$$(cat challenge/.gen/docker-image | cut -d ':' -f 2)
	if [[ "$${NEW_ID}" != $$(cat $@ 2>/dev/null) ]]; then
	  echo "$${NEW_ID}" > $@;
	fi
	docker tag "$${NEW_ID}" "kctf-chal-${CHALLENGE_NAME}"

challenge/.gen/docker-image: ../kctf-conf/base/nsjail-docker/.gen/docker-image .FORCE
	$(MAKE) -C challenge .gen/docker-image

.gen/kind-kubeconfig: .FORCE
	@command -v kind || (echo "error: kind not installed. Visit https://kind.sigs.k8s.io/" && false)
	kind get kubeconfig > $@

.gen/d4w-kubeconfig: .FORCE
	@command -v docker.exe || (echo "error: docker for windows not installed." && false)
	rm -f $@
	docker.exe context export default --kubeconfig $@

${CLUSTER_GEN}/image-tagged: .gen/challenge-image | .cluster-config
	IMAGE_ID="$$(cat .gen/challenge-image)"
	IMAGE_TAG="${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}:$${IMAGE_ID}"
	docker tag "$${IMAGE_ID}" "$${IMAGE_TAG}"
	echo -n "$${IMAGE_TAG}" > $@

${CLUSTER_GEN}/image-pushed: ${CLUSTER_GEN}/image-tagged
	IMAGE_TAG="$$(cat ${CLUSTER_GEN}/image-tagged)"
	if [ "${PUSH_TARGET}" == "REMOTE" ]; then
	  docker push "$${IMAGE_TAG}"
	fi
	if [ "${PUSH_TARGET}" == "KIND" ]; then
	  kind load docker-image "$${IMAGE_TAG}"
	fi
	touch $@

.gen/healthcheck-image: healthcheck/.gen/docker-image
	NEW_ID=$$(cat healthcheck/.gen/docker-image | cut -d ':' -f 2)
	if [[ "$${NEW_ID}" != $$(cat $@ 2>/dev/null) ]]; then
	  echo "$${NEW_ID}" > $@;
	fi
	docker tag "$${NEW_ID}" "kctf-healthcheck-${CHALLENGE_NAME}"

healthcheck/.gen/docker-image: ../kctf-conf/base/healthcheck-docker/.gen/docker-image .FORCE
	$(MAKE) -C healthcheck .gen/docker-image

${CLUSTER_GEN}/healthcheck-image-tagged: .gen/healthcheck-image | .cluster-config
	IMAGE_ID="$$(cat .gen/healthcheck-image)"
	IMAGE_TAG="${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}-healthcheck:$${IMAGE_ID}"
	docker tag "$${IMAGE_ID}" "$${IMAGE_TAG}"
	echo -n "$${IMAGE_TAG}" > $@

${CLUSTER_GEN}/healthcheck-image-pushed: ${CLUSTER_GEN}/healthcheck-image-tagged
	IMAGE_TAG="$$(cat ${CLUSTER_GEN}/healthcheck-image-tagged)"
	if [ "${PUSH_TARGET}" == "REMOTE" ]; then
		docker push "$${IMAGE_TAG}"
	fi
	if [ "${PUSH_TARGET}" == "KIND" ]; then
		kind load docker-image "$${IMAGE_TAG}"
	fi
	touch $@

../kctf-conf/base/nsjail-docker/.gen/docker-image: .FORCE
	$(MAKE) -C ${@D}/.. .gen/docker-image

../kctf-conf/base/healthcheck-docker/.gen/docker-image: .FORCE
	$(MAKE) -C ${@D}/.. .gen/docker-image

${CLUSTER_GEN}/remote-image: ${CLUSTER_GEN}/image-tagged .FORCE
	PUSHED_IMAGE="$$(cat ${CLUSTER_GEN}/image-tagged)"
	(kubectl get deployment/chal -o jsonpath='{.spec.template.spec.containers[?(@.name == "challenge")].image}' 2>/dev/null || echo -n "$${PUSHED_IMAGE}") > $@

${CLUSTER_GEN}/remote-healthcheck-image: ${CLUSTER_GEN}/healthcheck-image-tagged .FORCE
	PUSHED_IMAGE="$$(cat ${CLUSTER_GEN}/healthcheck-image-tagged)"
	REMOTE_TAG=$$(kubectl get deployment/chal -o jsonpath='{.spec.template.spec.containers[?(@.name == "healthcheck")].image}' 2>/dev/null || echo -n "$${PUSHED_IMAGE}")
	# if we previously deployed without a healthcheck, the output might be empty
	if [ -z "$${REMOTE_TAG}" ]; then
	  REMOTE_TAG="$${PUSHED_IMAGE}"
	fi
	echo -n "$${REMOTE_TAG}" > $@

.cluster-config:
	@if [ "${PROJECT}" = "CONFIGMISSING" ]; then
	@  echo 'error: cluster config not loaded. Run kctf-config-create or kctf-config-load'
	@  exit 1
	@fi
	kubectl config use-context "kctf_${PROJECT}_${ZONE}_${CLUSTER_NAME}" >&2
	kubectl config set-context --current --namespace="${CHALLENGE_NAME}" >&2
	kubectl create namespace "${CHALLENGE_NAME}" --dry-run=client -oyaml | kubectl apply -f - >&2
	kubectl patch ServiceAccount default --patch "automountServiceAccountToken: false" >&2
	mkdir -p ${CLUSTER_GEN} >&2

.FORCE:

define DEPLOYMENT_KUSTOMIZATION
commonLabels:
  app: "$(CHALLENGE_NAME)"
endef

define NETWORK_KUSTOMIZATION
bases:
- ../../../../kctf-conf/base/k8s/network

commonLabels:
  app: "$(CHALLENGE_NAME)"
endef

export DEPLOYMENT_KUSTOMIZATION
export NETWORK_KUSTOMIZATION
# This target doesn't live in a cluster specific directory since we need the name to be predictable
# To make up for it, we FORCE execution every time
.gen/k8s: ${CLUSTER_GEN}/remote-image ${MAYBE_REMOTE_HEALTHCHECK_IMAGE} .FORCE | .cluster-config
	mkdir -p $@

	# deployment
	mkdir -p $@/deployment
	# Use either deployment or deployment-with-healthcheck as base
	echo "bases:" >> "$@/deployment/kustomization.yaml"
ifeq ($(HEALTHCHECK), true)
	echo "- ../../../../kctf-conf/base/k8s/deployment-with-healthcheck" >> "$@/deployment/kustomization.yaml"
else
	echo "- ../../../../kctf-conf/base/k8s/deployment" >> "$@/deployment/kustomization.yaml" 
endif

	# set the image version to the currently deployed for challenge
	CHAL_IMAGE="$$(cat ${CLUSTER_GEN}/remote-image)"
	CHAL_IMAGE=($${CHAL_IMAGE//:/ })
	echo -e "\nimages:\n- name: challenge\n  newName: $${CHAL_IMAGE[0]}\n  newTag: \"$${CHAL_IMAGE[1]}\"" >> "$@/deployment/kustomization.yaml"

	# set the image version to the currently deployed for healthcheck
ifeq ($(HEALTHCHECK), true)
	HEALTHCHECK_IMAGE="$$(cat ${CLUSTER_GEN}/remote-healthcheck-image)"
	HEALTHCHECK_IMAGE=($${HEALTHCHECK_IMAGE//:/ })
	echo -e "- name: healthcheck\n  newName: $${HEALTHCHECK_IMAGE[0]}\n  newTag: \"$${HEALTHCHECK_IMAGE[1]}\"" >> "$@/deployment/kustomization.yaml"
endif

	# write the fixed customizations
	echo "$${DEPLOYMENT_KUSTOMIZATION}" >> "$@/deployment/kustomization.yaml"

	# network
	mkdir -p $@/network
	# write the fixed customizations
	echo "$${NETWORK_KUSTOMIZATION}" > "$@/network/kustomization.yaml"
	touch $@
