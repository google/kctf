#!/bin/bash

setpriv --init-groups --reset-env --reuid user --regid user --inh-caps=-all -- \
  nsjail --config /kctf/nsjail_base.cfg --config /home/user/nsjail.cfg --port 8081 -- \
    /home/user/chroot/node/run.sh &
