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
FROM gcr.io/kctf-docker/kctf-pwntools:latest AS pwntools
FROM ubuntu:20.04

RUN apt-get update && apt-get -yq --no-install-recommends install cpio openssl python3 && rm -rf /var/lib/apt/lists/*

RUN /usr/sbin/useradd --no-create-home -u 1000 user

COPY --from=pwntools /venv /venv
RUN mkdir -p /home/user/.pwntools-cache && echo never > /home/user/.pwntools-cache/update

COPY files/drop_privs /usr/bin/
COPY files/bypass_pow /usr/bin/
