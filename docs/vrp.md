# kCTF VRP Setup

We invite you to demonstrate an exploit against our kCTF demo cluster based on the challenges presented on this page. Successful demonstrations are eligible for rewards between [20,000 - 91,337 USD](https://security.googleblog.com/2022/02/roses-are-red-violets-are-blue-giving.html) as defined in [https://g.co/vrp](https://g.co/vrp). On top of this, exploiting our [new instances](#experimental-mitigations-challenge) are eligible up to 42,000 USD, increasing the total possible payout to [133,337 USD](https://security.googleblog.com/2022/08/making-linux-kernel-exploit-cooking.html).

[kCTF](https://github.com/google/kctf) is an open source infrastructure for CTF competitions. You can find details on how it works in the [kCTF documentation](https://google.github.io/kctf/introduction.html), but in short, it’s running on a hardened Kubernetes cluster with the following security features:

*   The OS and Kubernetes versions are upgraded automatically.
*   The nodes are running Container-Optimized OS.
*   Pod egress network access is restricted to public IPs only.
*   [Workload Identity](https://cloud.google.com/blog/products/containers-kubernetes/introducing-workload-identity-better-authentication-for-your-gke-applications) restricts access to service accounts and the metadata server in addition to the network policies.
*   Every connection to a challenge spawns a separate [nsjail](https://github.com/google/nsjail) sandbox to isolate players from each other.

At present, we’re interested in two attack scenarios against this infrastructure:

1. Breaking out of the nsjail sandbox as it would allow solving challenges in unintended ways.
2. Breaking the isolation that Kubernetes provides and accessing the flags of other challenges.

For this purpose, we set up two kCTF challenges with secret flags: “kctf” and “full-chain”. You can demonstrate a working exploit by leaking the flags of either of these.
You can find the code for the challenges
[here](https://github.com/google/google-ctf/tree/master/vrp).

![drawing showing the location of the flags](./images/flag-locations.png)


## kctf challenge

The “kctf” challenge is the only entry point to the cluster. You can connect to it via:


```
socat FILE:`tty`,raw,echo=0 TCP:kctf.vrp.ctfcompetition.com:1337
```

Alternative (for newer kernels)
```
socat FILE:`tty`,raw,echo=0 TCP:kctf.vrp2.ctfcompetition.com:1337
```


It will ask you to solve a proof-of-work and then gives you access to a bash running in a setup similar to the [kCTF pwn template challenge](https://github.com/google/kctf/tree/beta/dist/challenge-templates/pwn). The only difference is that the flag is not accessible inside of the nsjail sandbox and you will need to break out of the chroot in order to read it. You can observe the full source code [here](https://github.com/google/google-ctf/tree/master/vrp).

The details of the kernel of the VM can be read from `/etc/node-os-release`, and you can get the image of the VM following [this script](https://gist.github.com/sirdarckcat/568934df2b33a125b0b0f42a5366df8c) based on the output of `/etc/node-os-release`.


## full-chain challenge

The “full-chain” challenge is a challenge that runs a `while sleep` loop and doesn’t have any exposed ports. In order to get access to the flag, you will need to break out of the “kctf” challenge and break the pod isolation of the cluster.


## Flags

The flags are stored in Kubernetes [secrets](https://kubernetes.io/docs/concepts/configuration/secret/) and mounted to the filesystem of the two challenges at “/flag/flag”. They are of the format:


```
KCTF{$CHAL_NAME-$TIMESTAMP:$MAC}
```


As you can see, the flags include a timestamp and are rotated frequently.

## Experimental mitigations challenge

We’re [launching new instances](https://security.googleblog.com/2022/08/making-linux-kernel-exploit-cooking.html) to evaluate the latest Linux kernel stable image as well as new experimental mitigations in a custom kernel we've built. Rather than simply learning about the current state of the stable kernels, the new instances are used to ask the community to help us evaluate the value of both our latest and more experimental security mitigations.

The [mitigations](https://github.com/thejh/linux/blob/slub-virtual/MITIGATION_README) we've built attempt to tackle the following exploit primitives:
* Out-of-bounds write on slab
* Cross-cache attacks
* Elastic objects
* Freelist corruption

You can connect to the instance with the latest kernel without the patch via

```
nc kctf-mitigation.vrp.ctfcompetition.com 1337
```

And to the version patched with our experimental mitigations:

```
nc kctf-mitigation.vrp.ctfcompetition.com 31337
```

These instances are not based on the kCTF infrastructure (as they require running custom kernel version), instead they spin up a new QEMU VM on every new connection. As this is not a production-ready infrastructure, breaking the infrastructure itself (or e.g. using leaks from console) is not considered a valid submission.

|                | Upstream | Custom mitigation |
| -------------- | -------- | ----------------- |
| Kernel version | 5.19     | 5.19 - custom     |
| Kernel image   | [bzImage_upstream_5.19](https://storage.googleapis.com/kctf-vrp-public-files/bzImage_upstream_5.19) | [bzImage_mitigation_5.19](https://storage.googleapis.com/kctf-vrp-public-files/bzImage_mitigation_5.19) |
| Kernel config  | [bzImage_upstream_5.19_config](https://storage.googleapis.com/kctf-vrp-public-files/bzImage_upstream_5.19_config) | [bzImage_mitigation_5.19_config](https://storage.googleapis.com/kctf-vrp-public-files/bzImage_mitigation_5.19_config) |
| Source code    | [3d7cb6b](https://github.com/thejh/linux/tree/3d7cb6b04c3f3115719235cc6866b10326de34cd) | [c02401c](https://github.com/thejh/linux/tree/c02401c87a2d84efb47c4354400a9ad17d7b6436) |
| Port           | 1337     | 31337             |
| Reward (base)  | $21000   | $42000            |

Please also note that although we are trying to keep up-to-date with the latest kernel version, these instances may sometimes be outdated.

### Submission

We want to avoid learning about unfixed vulnerabilities, so the process to submit reports is:
  1. Test your exploit - we recommend you to test it locally first, and run a GKE cluster to debug.
  2. If it is a 0day (there's no patch for it on [linus master branch](https://github.com/torvalds/linux/tree/master) yet), then send us a checksum of your working exploit to our form [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform). You won't share any technical details about the vulnerability, you will just record the fact you found something (as we only reward the first person that writes an exploit for a given bug, we use it to resolve the timing in case of an exploit collision). Make sure to submit the exploit checksum **before** there's a public patch and to submit the full exploit **within a week** after the patch is public. If you take longer than a week, we might issue the reward to someone else.
  3. For 1days or once there is a public patch, test your exploit it on the [lab environment](#kctf-challenge). If you have troubles let us know in [#kctf](https://discord.gg/V8UqnZ6JBG) and we'll help you figure out any problems.
  4. Once you get the flag, send it together with the patch and the exploit [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform).
  5. To increase the timely sharing of new techniques with the community, we are also now requiring that the exploits that receive innovation bonus get publicly documented within a month, otherwise we may publish it.

### Notes

We want to encourage the community to help research vulnerabilities such as those [found by Syzkaller](https://syzkaller.appspot.com/upstream#open), but which are still unfixed since they have not been shown to be exploitable. As such:



*   The person that develops the exploit and receives the reward might not be the same as the person that discovered or patched the vulnerability.
*   It's ok to use 1-day exploits against the lab environment using publicly known vulnerabilities that exploit the patch gap between the time when a patch is announced and the lab environment is updated, however we will only issue a single reward per vulnerability.


**When we receive an exploit for a fixed vulnerability we'll add details [here](https://docs.google.com/spreadsheets/d/e/2PACX-1vS1REdTA29OJftst8xN5B5x8iIUcxuK6bXdzF8G1UXCmRtoNsoQ9MbebdRdFnj6qZ0Yd7LwQfvYC2oF/pubhtml).**

In case of questions or suggestions, you can reach us in [#kctf](https://discord.gg/V8UqnZ6JBG).
