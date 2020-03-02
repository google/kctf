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
FROM kctf-nsjail-chroot:latest AS chroot
FROM kctf-nsjail-bin:latest AS bin
FROM ubuntu:19.10

RUN apt-get update && apt-get install -y uidmap libprotobuf17 libnl-route-3-200
RUN apt-get upgrade -y

COPY --from=bin /usr/bin/nsjail /usr/bin/nsjail
COPY --from=chroot /chroot /chroot

RUN mkdir /chroot/config
RUN mkdir /chroot/secrets

COPY files/k8s_nsjail_setup.sh /usr/bin/
COPY files/proof_of_work/pow.py /chroot/usr/bin/
COPY files/proof_of_work/maybe_pow.sh /chroot/usr/bin/
