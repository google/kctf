# kCTF
[![GKE Deployment](https://github.com/google/kctf/workflows/GKE%20Deployment/badge.svg?branch=master)](https://github.com/google/kctf/actions?query=workflow%3A%22GKE+Deployment%22)

kCTF is a Kubernetes-based infrastructure for CTF competitions.

## Prerequisites

* [gcloud](https://cloud.google.com/sdk/install)
* [docker](https://docs.docker.com/install/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.17+

## Getting Started

The following documentation resources are available to help you get started:

* [Local Testing Walkthrough](local-testing.md) – A quick start guide showing you how to build and test challenges locally.
* [kCTF in 8 Minutes](introduction.md) – A quick 8-minute summary of what kCTF is and how it interacts with Kubernetes.
* [Google Cloud Walkthrough](google-cloud.md) – Once you have everything up and running, try deploying to Google Cloud. 
* [Troubleshooting](troubleshooting.md) – Help with fixing broken challenges.
* [DNS Setup](dns.md) – Information on setting up DNS for the CTF.
* [Security Threat Model](security-threat-model.md) – Security considerations regarding kCTF including information on assets, risks, and potential attackers.

## Samples

The [samples](https://github.com/google/kctf/tree/master/samples) directory contains two examples:
* A web challenge that acts like an XSS bot.
* A web challenge that acts like a vulnerable PHP application with support for sessions and file uploads.

