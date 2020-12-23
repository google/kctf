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
FROM ubuntu:20.04 as chroot

RUN /usr/sbin/useradd --no-create-home -u 1000 user

FROM gcr.io/kctf-docker/challenge

COPY --from=chroot / /chroot

COPY flag /chroot/

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -yq --no-install-recommends tzdata \
    && ln -fs /usr/share/zoneinfo/Europe/Berlin /etc/localtime \
    && dpkg-reconfigure --frontend noninteractive tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -yq --no-install-recommends apache2 && rm -rf /var/lib/apt/lists/*

RUN service apache2 start

RUN ln -s /etc/apache2/mods-available/proxy.load /etc/apache2/mods-enabled/
RUN ln -s /etc/apache2/mods-available/proxy_http.load /etc/apache2/mods-enabled/
COPY apache2-nsjail-others.conf /etc/apache2/conf-enabled/

RUN chroot /chroot apt-get update \
    && chroot /chroot apt-get install -yq --no-install-recommends \
       curl ca-certificates socat gnupg lsb-release software-properties-common \
    && rm -rf /chroot/var/lib/apt/lists/*

RUN chroot /chroot curl -sSL https://deb.nodesource.com/gpgkey/nodesource.gpg.key | chroot /chroot apt-key add - \
    && (echo "deb https://deb.nodesource.com/node_10.x $(chroot /chroot lsb_release -s -c) main";\
        echo "deb-src https://deb.nodesource.com/node_10.x $(chroot /chroot lsb_release -s -c) main") \
       > /chroot/etc/apt/sources.list.d/nodesource.list \
    && chroot /chroot add-apt-repository universe \
    && chroot /chroot apt-get update \
    && chroot /chroot apt-get install -yq --no-install-recommends nodejs socat \
    && rm -rf /chroot/var/lib/apt/lists/*

RUN mkdir -p /chroot/home/user
RUN mkdir -p /chroot/home/user/chroot
COPY chroot /home/user/chroot
COPY node /home/user/node
COPY nsjail.cfg /home/user/nsjail.cfg

CMD kctf_setup \
    && /home/user/node/start.sh \
    && mount -t tmpfs none /var/run/apache2 \
    && bash -c 'source /etc/apache2/envvars && APACHE_RUN_USER=user APACHE_RUN_GROUP=user /usr/sbin/apache2 -D FOREGROUND'
