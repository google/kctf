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
FROM ubuntu:19.10

RUN apt-get update && apt-get install -y --no-install-recommends debootstrap

RUN debootstrap --variant minbase --include python3 eoan /chroot
RUN chroot /chroot /usr/sbin/useradd --no-create-home -u 1000 user
RUN /usr/sbin/useradd --no-create-home -u 1000 user
