# Quickstart guide to writing a challenge:

* A docker image will be built from `challenge/image`. For the simplest challenges, replacing `challenge/image/chal` will be enough.
* Edit `challenge/image/Dockerfile` to change the commandline or which files should get included.
* Try the challenge locally with `make test-docker`.
* If you have a cluster ready, deploy the challenge with `make start`.
  * To access it, create a port forward with `make port-forward` and connect via `nc localhost PORT` using the printed port.
* Check out `make <tab>` for more commands.

## healthcheck

* The healthcheck needs to run a webserver on port 45281 that responds to /healthz requests.
 * 200 if the challenge is healthy, otherwise an error (e.g. 400).
 * You can find an example webserver in `healthcheck/image/healthz.py`.
* The base image comes preloaded with pwntools.
* If your exploit is written in python, just drop it in `healthcheck/image/doit.py`.

# Directory layout:

## /config

The config directory holds a few configuration files:

* chal.conf: Configure if the challenge should be deployed to the cluster, if it's publicly accessible and if it has a healthcheck.
* pow.conf: Contains the difficulty of the proof-of-work, 0 means disabled.
* advanced: The kubernetes config used to deploy the challenge. E.g. you can add a tmpfs here or change the port.

## /challenge

The challenge binary and anything else challenge related goes in here.

* image/: Dockerfile and files of the challenge image.
* Makefile: You can edit this file if you need to run code before building the docker image, e.g. for building the challenge from source.

## /healthcheck

The healthcheck directory is optional. If you don't want to write a healthcheck, feel free to delete it. You really should have healthchecks though :).

* image/: Dockerfile and files of the healthcheck.
* Makefile: same as for the challenge directory.

# API contract

Here are the requirements what this directory must look like to work with kCTF:

* You can do pretty much whatever you want in the challenge dir but:
  * You need to have a Makefile with the .gen/docker-image target that builds a docker image.
  * We strongly recommend to use nsjail in all challenges.
* Your challenge will receive connections on port 1337.
* The healthcheck directory is optional.
  * If it exists, the image should run a webserver on port 45281 and respond to /healthz requests.
* Changes to config/advanced might not be compatible with future versions of kCTF.
