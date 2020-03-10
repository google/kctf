#!/bin/bash

setpriv --init-groups --reset-env --reuid user --regid user --inh-caps=-all \
  nsjail --config /config/nsjail.cfg --port 8081 -- \
  /home/user/chroot/node/run.sh &
