FROM ubuntu:24.04

RUN apt-get update && apt-get install -y wget fuse
RUN wget -q https://github.com/GoogleCloudPlatform/gcsfuse/releases/download/v1.4.2/gcsfuse_1.4.2_amd64.deb && dpkg -i gcsfuse_1.4.2_amd64.deb
RUN mkdir -p /mnt/disks/gcs

CMD test -f /config/gcs_bucket &&\
    gcsfuse --foreground --debug_fuse --debug_gcs --stat-cache-ttl 0 -o allow_other --file-mode 0777 --dir-mode 0777 --uid 1000 --gid 1000 "$(cat /config/gcs_bucket)" /mnt/disks/gcs
