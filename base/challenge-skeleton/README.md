# Quickstart guide to writing a challenge

The basic steps when preparing a challenge are:

* A Docker image is built from `challenge/image`. For the simplest challenges, replacing `challenge/image/chal` is sufficient.
* Edit `challenge/image/Dockerfile` to change the commandline or the files you want to include.
* Try the challenge locally with `make test-docker`.
* If you have prepared a cluster, deploy the challenge with `make start`.
  * To access the challenge, create a port forward with `make port-forward` and connect to it via `nc localhost PORT` using the printed port.
* Check out `make <tab>` for more commands.

## Directory layout

The following directories are available:

### /config

The `config` directory holds a few configuration files:

* `chal.conf`: Configure this file if you are deploying the challenge to the cluster, if it's publicly accessible, and/or if it has a healthcheck.
* `pow.conf`: Contains the difficulty of the proof-of-work, 0 means disabled.
* `advanced`: The Kubernetes config used to deploy the challenge. E.g. you can add a tmpfs here or change the port.

### /challenge

The `challenge` directory contains the challenge binary and anything else related to the challenge:

* `image/`: Dockerfile and files of the challenge image.
* `Makefile`: Edit this file if you need to run code before building the Docker image, e.g. for building the challenge from source.

### /healthcheck

The `healthcheck` directory is optional. If you don't want to write a healthcheck, feel free to delete the directory. However, we strongly recommend that you implement a healthcheck :).

* `image/`: Dockerfile and files of the healthcheck.
* `Makefile`: Edit this file if you need to build any custom files before running your healthcheck.


#### Healthcheck

Important details regarding healthchecks:

* Run the healthcheck on a webserver on port 45281 that responds to `/healthz` requests.
 * The healthcheck returns 200 if the challenge is healthy, otherwise an error (e.g. 400).
 * You can find an example webserver in `healthcheck/image/healthz.py`.
* The base image comes preloaded with pwntools.
* If your exploit is written in Python, drop it in `healthcheck/image/doit.py`.

## API contract

Ensure your setup fulfills the following requirements to ensure it works with kCTF:

* In your cmdline, call `kctf_setup` as the first command.
* You can do pretty much whatever you want in the `challenge` directory but:
  * You need to have a `Makefile` with the `.gen/docker-image` target that builds a Docker image.
  * We strongly recommend using nsjail in all challenges.
* Your challenge receives connections on port 1337.
* The healthcheck directory is optional.
  * If it exists, the image should run a webserver on port 45281 and respond to `/healthz` requests.
  
Note: Changes to `config/advanced` might not be compatible with future versions of kCTF.
