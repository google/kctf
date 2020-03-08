#!/bin/bash

mkfifo /tmp/node
node /home/user/chroot/node/app.js >/tmp/node &
head -n 1 /tmp/node >/dev/null
socat - TCP:127.0.0.1:8080
