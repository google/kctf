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

.PHONY: start stop docker ip status logs healthcheck-logs ssh healthcheck-ssh port-forward test-docker healthcheck-test-docker test-kind test-d4w .test-local clean .deploy .cluster-config .FORCE

SHELL := bash
.ONESHELL:
.SHELLFLAGS = -e -c

PUSH_TARGET="REMOTE"

export REGISTRY:=CONFIGMISSING
export PROJECT:=CONFIGMISSING
export CLUSTER_NAME:=CONFIGMISSING
export ZONE:=CONFIGMISSING

-include $(HOME)/.config/kctf/cluster.conf
KUBECONFIG=$(HOME)/.config/kctf/kube.conf
export KUBECONFIG

export CHALLENGE_NAME:=$(shell basename ${CURDIR})
CLUSTER_GEN=.gen/${PROJECT}_${ZONE}_${CLUSTER_NAME}
export REMOTE_IMAGE:=${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}
export REMOTE_HEALTHCHECK_IMAGE:=${REGISTRY}/${PROJECT}/${CHALLENGE_NAME}-healthcheck

HEALTHCHECK_ENABLED:=$(shell yq read challenge.yaml 'spec.healthcheck.enabled')

docker: .gen/challenge-image

start: .deploy

stop: .cluster-config
	kubectl delete namespace ${CHALLENGE_NAME}

ip: .cluster-config
	@LB_IP=""
	@while [ -z "$${LB_IP}" ]; do
	@  LB_IP=$$(kubectl get "service/${CHALLENGE_NAME}" -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
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
	kubectl exec deployment/${CHALLENGE_NAME} -c challenge -it -- /bin/bash

healthcheck-ssh: .cluster-config
	kubectl exec deployment/${CHALLENGE_NAME} -c healthcheck -it -- /bin/bash

port-forward: .cluster-config
	kubectl port-forward deployment/${CHALLENGE_NAME} :1337 &

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
	kubectl patch deployment ${CHALLENGE_NAME}  --type json -p='[{"op": "remove", "path": "/spec/template/metadata/annotations/container.apparmor.security.beta.kubernetes.io~1challenge"}]' --kubeconfig=${LOCAL_KUBECONFIG}

clean:
	rm -R .gen/* || true
	rm -R challenge/.gen/* || true
	rm -R healthcheck/.gen/* || true

.deploy: .deployment .cluster-config

.deployment: challenge.yaml .FORCE | .cluster-config
	kubectl apply -f challenge.yaml

challenge.yaml: ${CLUSTER_GEN}/challenge-image-pushed ${CLUSTER_GEN}/healthcheck-image-pushed
	yq write --inplace challenge.yaml 'spec.image' $$(cat ${CLUSTER_GEN}/challenge-image-pushed)
ifeq ($(HEALTHCHECK_ENABLED), true)
	yq write --inplace challenge.yaml 'spec.healthcheck.image' $$(cat ${CLUSTER_GEN}/healthcheck-image-pushed)
endif

.gen/challenge-image: challenge/.gen/docker-image
	NEW_ID=$$(cat challenge/.gen/docker-image)
	if [[ "$$NEW_ID" = sha256:* ]]; then
	  NEW_ID=$$(echo "$$NEW_ID" | cut -d ':' -f 2)
	fi
	echo "$${NEW_ID}" > $@;

challenge/.gen/docker-image: ../kctf-conf/base/nsjail-docker/.gen/docker-image .FORCE
	$(MAKE) -C challenge .gen/docker-image

.gen/kind-kubeconfig: .FORCE
	@command -v kind || (echo "error: kind not installed. Visit https://kind.sigs.k8s.io/" && false)
	kind get kubeconfig > $@

.gen/d4w-kubeconfig: .FORCE
	@command -v docker.exe || (echo "error: docker for windows not installed." && false)
	rm -f $@
	docker.exe context export default --kubeconfig $@

${CLUSTER_GEN}/challenge-image-pushed: .gen/challenge-image | .cluster-config
	IMAGE_ID="$$(cat .gen/challenge-image)"
	if [[ "$${IMAGE_ID}" = "$${REMOTE_IMAGE}:"* ]]; then
	  IMAGE_TAG="$${IMAGE_ID}"
	else
	  IMAGE_TAG="${REMOTE_IMAGE}:$${IMAGE_ID}"
	  docker tag "$${IMAGE_ID}" "$${IMAGE_TAG}"
	  if [ "${PUSH_TARGET}" == "REMOTE" ]; then
	    docker push "$${IMAGE_TAG}"
	  fi
	fi
	if [ "${PUSH_TARGET}" == "KIND" ]; then
	  kind load docker-image "$${IMAGE_TAG}"
	fi
	echo -n "$${IMAGE_TAG}" > $@

.gen/healthcheck-image: healthcheck/.gen/docker-image
	NEW_ID=$$(cat healthcheck/.gen/docker-image)
	if [[ "$$NEW_ID" = sha256:* ]]; then
	  NEW_ID=$$(echo "$$NEW_ID" | cut -d ':' -f 2)
	fi
	echo "$${NEW_ID}" > $@;

healthcheck/.gen/docker-image: ../kctf-conf/base/healthcheck-docker/.gen/docker-image .FORCE
	$(MAKE) -C healthcheck .gen/docker-image

${CLUSTER_GEN}/healthcheck-image-pushed: .gen/healthcheck-image | .cluster-config
ifeq ($(HEALTHCHECK_ENABLED), true)
	IMAGE_ID="$$(cat .gen/healthcheck-image)"
	if [[ "$${IMAGE_ID}" = "$${REMOTE_HEALTHCHECK_IMAGE}:"* ]]; then
	  IMAGE_TAG="$${IMAGE_ID}"
	else
	  IMAGE_TAG="${REMOTE_HEALTHCHECK_IMAGE}:$${IMAGE_ID}"
	  docker tag "$${IMAGE_ID}" "$${IMAGE_TAG}"
	  if [ "${PUSH_TARGET}" == "REMOTE" ]; then
	    docker push "$${IMAGE_TAG}"
	  fi
	fi
	if [ "${PUSH_TARGET}" == "KIND" ]; then
	  kind load docker-image "$${IMAGE_TAG}"
	fi
	echo -n "$${IMAGE_TAG}" > $@
endif

../kctf-conf/base/nsjail-docker/.gen/docker-image: .FORCE
	$(MAKE) -C ${@D}/.. .gen/docker-image

../kctf-conf/base/healthcheck-docker/.gen/docker-image: .FORCE
	$(MAKE) -C ${@D}/.. .gen/docker-image

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
