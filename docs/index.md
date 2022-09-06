# kCTF

kCTF is CTF infrastructure written on top of Kubernetes. It allows you to manage, deploy, and sandbox CTF challenges.

## [Try it – it takes 5 minutes to get started](local-testing.md)

> We built kCTF to help CTF organizers have great CTF infrastructure tooling, and to give challenge authors an environment where they can build and test their challenges without having to learn about Kubernetes, while also allowing Kubernetes experts to use all the features they are accustomed to.


| **Versatile** | **Specialized** | **Secure** |
|:--------------|:----------------|:-----------|
| A simple challenge can be configured in a single Dockerfile. More complex deployments can use all Kubernetes features. | Common CTF needs are provided as services (proof of work, healthchecks, DNS and SSL certificates, network fileshare). | kCTF was built by Google's Security Team with the goal of hosting untrusted, vulnerable applications as part of CTF competitions. |

## Preview
You can have a local challenge running in 5 minutes.

[<img src="https://user-images.githubusercontent.com/33089/111788876-df83fe80-88c0-11eb-8485-f147bc23d7ca.gif" width="600">](https://asciinema.org/a/sePuQKLBHaO3JOtQj9gWayWvU)


See the [Local Testing Walkthrough](local-testing.md) to get started with a local challenge.

## Google Cloudshell Codelab
To try this out on a real server, try out the Google Cloudshell codelab.

[![Open in Cloudshell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/google/kctf&tutorial=docs/google-cloud.md&shellonly=true)

## Online demo
We have a fake challenge running, you can see what an isolated challenge would look like by connecting directly to:
```bash
nc kctf.vrp.ctfcompetition.com 1337
```

If you are able to break out of it, you can [earn up to $133,337 USD](vrp.md).

**Tip**: Execute `bash -i 2>&1` to get an interactive bash session.

## Available Documentation

* [Local Testing Walkthrough](local-testing.md) – A quick start guide showing you how to build and test challenges locally.
* [kCTF in 8 Minutes](introduction.md) – A quick 8-minute summary of what kCTF is and how it interacts with Kubernetes.
* [Google Cloud Walkthrough](google-cloud.md) – Once you have everything up and running, try deploying to Google Cloud.
* [Custom Domains](custom-domains.md) – How to add custom domains for your challenges.
* [Troubleshooting](troubleshooting.md) – Help with fixing broken challenges.
* [CTF playbook](ctf-playbook.md) – How to set up your cluster and challenges to scale during a CTF.
* [Security Threat Model](security-threat-model.md) – Security considerations regarding kCTF including information on assets, risks, and potential attackers.
* [kCTF VRP Setup](vrp.md) – Demonstrate an exploit against our kCTF demo cluster based on the challenges presented on this page.
