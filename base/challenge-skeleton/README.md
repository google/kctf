# Quickstart guide to writing a challenge:

* A docker image will be built from challenge/image. For the simplest challenges, replacing challenge/image/chal will be enough.
* To change what files get included or the commandline, edit challenge/image/Dockerfile. TODO: add commandline to dockerfile
* Try the challenge locally with `make test-docker`.
* If you have a cluster ready, deploy the challenge with `make start` and create a port forward with `make port-forward`.
* Try `make <tab>` to see more commands.

## healthcheck

* The healthcheck needs to run a webserver on port TODO that responds to /healthz requests.
 * 200 if the challenge is healthy, otherwise an error (e.g. 400).
 * You can find an example webserer in healthcheck/image/healthz.py.
* The base image comes preloaded with pwntools.
* If your exploit is written in python, just drop it in healthcheck/image/doit.py.

# Directory layout:

## /config

The config directory holds a few configuration files:

* chal.conf: Configure if the challenge should be deployed to the cluster if it's publicly accessible and if it has a healthcheck.
* pow.conf: Change the proof-of-work difficulty in this file, 0 means disabled.
* advanced: The kubernetes config used to deploy the challenge. E.g. add a tmpfs here or change the port.

## /challenge

The challenge binary and anything else challenge related goes in here.

* image/: Dockerfile and files of the challenge image.
* Makefile: You can edit this file if you need to run code before building the docker image, e.g. for building the challenge from source.

## /healthcheck

The healthcheck directory is optional. If you don't want to write a healthcheck, feel free to delete it. You really should though :).

* image/: Dockerfile and files of the healthcheck.
* Makefile: same as for the challenge directory.

# API contract

Here are the requirements what this directory must look like to work with kCTF:
* You can do pretty much whatever you want in the challenge dir but:
  * You need to have a Makefile with the .gen/docker-image that builds a docker image
  * We strongly recommend to use nsjail in all challenges.
* Your challenge will receive connections on port 1337.
* The healthcheck directory is optional.
  * If it exists, the image should run a webserver on port TODO and respond to /healthz requests.
* Changes to config/advanced might not be compatible with future versions of kCTF.
