# kCTF
[![GKE Deployment](https://github.com/google/kctf/workflows/GKE%20Deployment/badge.svg?branch=master)](https://github.com/google/kctf/actions?query=workflow%3A%22GKE+Deployment%22)

kCTF is a Kubernetes-based infrastructure for CTF competitions.

## Prerequisites

* [gcloud](https://cloud.google.com/sdk/install)
* [docker](https://docs.docker.com/install/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.17+

## Getting Started

If you want to quickly get started, **follow the [local testing](docs/local-testing.md) walkthrough**, this will show you how to build and test challenges locally.

If you want to read more about how kCTF works, take a look at [kCTF in 8 minutes](docs/introduction.md): A quick 8 minutes summary of what is kCTF and how it interacts with Kubernetes.

Once you have everything up and running, you can try deploying to Google Cloud. Follow the [Google Cloud walkthrough](docs/google-cloud.md).

You can also find a [troubleshooting playbook](docs/troubleshooting.md) (for fixing broken challenges), and a guide for [setting up DNS](docs/dns.md) for the CTF.

## Samples

In the [samples](samples) directory you can find a couple example web challenges. One of them is a challenge that acts as an XSS bot, and the other is a challenge that acts as a vulnerable PHP application with support for sessions and file uploads.

