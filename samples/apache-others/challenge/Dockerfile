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
FROM kctf-nsjail

RUN apt-get update && apt-get install -yq --no-install-recommends apache2 && rm -rf /var/lib/apt/lists/*

RUN service apache2 start

RUN ln -s /etc/apache2/mods-available/proxy.load /etc/apache2/mods-enabled/
RUN ln -s /etc/apache2/mods-available/proxy_http.load /etc/apache2/mods-enabled/
RUN ln -s /config/apache2-nsjail-others.conf /etc/apache2/conf-enabled/

RUN echo nameserver 8.8.8.8 > /chroot/etc/resolv.conf

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

RUN mkdir -p /chroot/home/user && chown user:user /chroot/home/user
COPY files /home/user/
RUN chown -R user:user /home/user

CMD /usr/bin/k8s_nsjail_setup.sh \
    && /home/user/node/start.sh \
    && bash -c 'source /etc/apache2/envvars && APACHE_RUN_USER=user APACHE_RUN_GROUP=user /usr/sbin/apache2 -D FOREGROUND'
