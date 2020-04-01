kCTF is a Kubernetes-based infrastructure for CTF competitions.

## Documentation

The following documentation resources are available:

* [Local testing walthrough](docs/local-testing.md) – A quick start guide showing you how to build and test challenges locally.
* [kCTF in 8 minutes](docs/introduction.md) – A quick 8-minute summary of what kCTF is and how it interacts with Kubernetes.
* [Google Cloud walkthrough](docs/google-cloud.md) – Once you have everything up and running, try deploying to Google Cloud. 
* [Troubleshooting](docs/troubleshooting.md) – Help with fixing broken challenges.
* [DNS setup](docs/dns.md) – Information on setting up DNS for the CTF.

## Prerequisites

* [gcloud](https://cloud.google.com/sdk/install)
* [docker](https://docs.docker.com/install/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.17+

## Samples

The [samples](samples) directory contains two examples:
* A web challenge that acts like an XSS bot.
* A web challenge that acts like a vulnerable PHP application with support for sessions and file uploads.
