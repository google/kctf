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

# build nsjail first
FROM ubuntu:20.04 as nsjail

ENV BUILD_PACKAGES build-essential git protobuf-compiler libprotobuf-dev bison flex pkg-config libnl-route-3-dev ca-certificates
ENV NSJAIL_COMMIT 4be95952340bd1339889174d3a9158d636985813

RUN apt-get update \
    && apt-get install -yq --no-install-recommends $BUILD_PACKAGES \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/google/nsjail.git \
    && cd /nsjail && git checkout $NSJAIL_COMMIT && make -j && cp nsjail /usr/bin/ \
    && rm -R /nsjail && apt-get remove --purge -y $BUILD_PACKAGES $(apt-mark showauto)

# challenge image
FROM ubuntu:20.04

RUN apt-get update \
    && apt-get install -yq --no-install-recommends build-essential python3-dev python3.8 python3-pip libgmp3-dev libmpc-dev uidmap libprotobuf17 libnl-route-3-200 wget netcat ca-certificates socat \
    && rm -rf /var/lib/apt/lists/*

RUN /usr/sbin/useradd --no-create-home -u 1000 user

COPY --from=nsjail /usr/bin/nsjail /usr/bin/nsjail

# gmpy2 and ecdsa used by the proof of work
RUN python3 -m pip install ecdsa gmpy2

# we need a clean proc to allow nsjail to remount it in the user namespace
RUN mkdir /kctf
RUN mkdir -p /kctf/.fullproc/proc
RUN chmod 0700 /kctf/.fullproc

COPY kctf_setup /usr/bin/
COPY kctf_drop_privs /usr/bin/
COPY kctf_pow /usr/bin/
COPY pow.py /kctf/
