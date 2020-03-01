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
.PHONY: docker

# Use `make docker NSJAIL_DOCKER=-nodeps` to use remote cache.
NSJAIL_DOCKER=

docker: .gen/docker-image${DOCKER_IMAGE}

.gen/docker-image: | .gen/nsjail-docker .gen/chroot-docker .gen/docker-image-nodeps
	cat .gen/docker-image-nodeps > $@

.gen/docker-image-nodeps: Dockerfile $(shell find files)
	docker build -t "kctf-nsjail" .
	echo $$(docker image ls "kctf-nsjail" -q) > $@

.gen/nsjail-docker: nsjail.Dockerfile
	docker build -t docker.pkg.github.com/google/kctf/kctf-nsjail:latest -f nsjail.Dockerfile .
	echo $$(docker image ls "docker.pkg.github.com/google/kctf/kctf-nsjail:latest" -q) > $@

.gen/chroot-docker: chroot.Dockerfile
	docker build -t docker.pkg.github.com/google/kctf/kctf-chroot:latest -f chroot.Dockerfile .
	echo $$(docker image ls "docker.pkg.github.com/google/kctf/kctf-chroot:latest" -q) > $@
