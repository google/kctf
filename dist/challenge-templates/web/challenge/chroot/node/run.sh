#!/bin/bash

# Create a pipe
mkfifo /tmp/pipe

# Start node web server
node /home/user/chroot/node/app.js >/tmp/pipe &

# This will wait for /tmp/pipe to be written to
status=$(head -n 1 /tmp/pipe)

# Proxy stdin/stdout to web server
socat - TCP:127.0.0.1:8080
