# kCTF

kCTF is CTF infrastructure written on top of Kubernetes. It allows you to manage, deploy and sandbox CTF challenges.

## [Try it - it takes 5 minutes to get started](local-testing.md)

> We built kCTF to help CTF organizers have great CTF infrastructure tooling, and to give challenge authors an environment so they can build and test their challenges without having to learn about Kubernetes, while also allowing Kubernetes experts to use all the features they are accostumed to.


| **Versatile** | **Specialized** | **Secure** |
|:--------------|:----------------|:-----------|
| A simple challenge can be configured in single Dockerfile. More complex deployments can use all of kubernetes features. | Common CTF needs are provided as services (proof of work, healthchecks, DNS and SSL certificates, network fileshare). | kCTF was built by Google's Security Team with the goal of host untrusted vulnerable applications as part of CTF competitions. |

## Preview
You can have a local challenge running in 5 minutes.

[<img src="https://user-images.githubusercontent.com/33089/111788876-df83fe80-88c0-11eb-8485-f147bc23d7ca.gif" width="600">](https://asciinema.org/a/sePuQKLBHaO3JOtQj9gWayWvU)


[Get started](local-testing.md)

## Google Cloudshell Codelab
To try this out in a real server try out Google Cloudshell codelab.

[![Open in Cloudshell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/google/kctf&tutorial=docs/google-cloud.md&shellonly=true)

## Online demo
We have a fake challenge running, you can see what an isolated challenge would look like by connecting directly to:
```bash
nc kctf.vrp.ctfcompetition.com 1337
```

If you are able to break out of it you can [earn up to $10,000 USD](vrp.md).

**Tip**: Execute `bash -i 2>&1` to get an interactive bash session.
