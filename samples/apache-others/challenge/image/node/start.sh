#!/bin/bash

mount -t tmpfs none /var/log/apache2

kctf_drop_privs nsjail --config /home/user/nsjail.cfg --port 8081 -- \
  /home/user/chroot/node/run.sh &
