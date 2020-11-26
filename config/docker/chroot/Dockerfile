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
FROM ubuntu:20.04

RUN apt-get update \
    && apt-get install -yq --no-install-recommends debootstrap \
    && rm -rf /var/lib/apt/lists/* \
    && debootstrap --include python3.8 --variant minbase focal /chroot \
    && echo nameserver 8.8.8.8 > /chroot/etc/resolv.conf \
    && chroot /chroot /usr/sbin/useradd --no-create-home -u 1000 user \
    && apt-get remove --purge -y debootstrap $(apt-mark showauto)
