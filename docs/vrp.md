# kCTF VRP Setup

We invite you to demonstrate an exploit against our kCTF demo cluster based on the challenges presented on this page. Successful demonstrations are eligible for rewards between ~5,000 to 10,000 USD~[31,337 - 50,337 USD](https://security.googleblog.com/2021/11/trick-treat-paying-leets-and-sweets-for.html) as defined in https://g.co/vrp.

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

### Submission

We want to avoid learning about unfixed vulnerabilities, so the process to submit reports is:
  1. Test your exploit - we recommend you to test it locally first, and run a GKE cluster to debug.
  2. If it is a 0day (there's no patch for it on [linus master branch](https://github.com/torvalds/linux/tree/master) yet), then send us a checksum of your working exploit to our form [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform). You won't share any technical details about the vulnerability, you will just record the fact you found something (as we only reward the first person that writes an exploit for a given bug, we use it to resolve the timing in case of an exploit collision). Make sure to submit the exploit checksum **before** there's a public patch and to submit the full exploit **within a week** after the patch is public. If you take longer than a week, we might issue the reward to someone else.
  3. For 1days or once there is a public patch, test your exploit it on the [lab environment](#kctf-challenge). If you have troubles let us know in [#kctf](https://discord.gg/V8UqnZ6JBG) and we'll help you figure out any problems.
  4. Once you get the flag, send it together with the patch and the exploit [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform).

### Notes

We want to encourage the community to help research vulnerabilities such as those [found by Syzkaller](https://syzkaller.appspot.com/upstream#open), but which are still unfixed since they have not been shown to be exploitable. As such:



*   The person that develops the exploit and receives the reward might not be the same as the person that discovered or patched the vulnerability.
*   It's ok to use 1-day exploits against the lab environment using publicly known vulnerabilities that exploit the patch gap between the time when a patch is announced and the lab environment is updated, however we will only issue a single reward per vulnerability.


When we receive an exploit for a fixed vulnerability we'll add details here.

|Form Received|Exploit Received|Patch|References|
|:--:|:--:|:--:|:--:|
|2021-12-14|2021-12-14|[Patch](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=3b0462726e7ef281c35a7a4ae33e93ee2bc9975b)|[checksum](#93d608415be627643697d554c00ce93c9ea434619a8ea220e1c1cc21902930cf)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=1bef50bdd9622a1969608d1090b2b4a588d0c6ac)<br/>[CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=2021-4154)|
|2021-12-24|2021-12-24|[Patch](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=3b0462726e7ef281c35a7a4ae33e93ee2bc9975b)|[checksum](#021590cc2ea40a4134d06a87d9e66b4885f41915efce8a02b2498ff1c2f3170d)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=1bef50bdd9622a1969608d1090b2b4a588d0c6ac)<br/>[CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=2021-4154)|
|2021-12-27|?|?|[checksum](#dfad77dfb3fdd5c640834f548cba6c90bacf8f14a5e7eb9a980b61ea084fb03b)|
|2022-01-05|2022-01-05|[Patch](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=ec6af094ea28f0f2dda1a6a33b14cd57e36a9755)|[checksum](#b3b0e71c36081fdf58bc63fafec32007e45ea1f4a7ca4b7d7432b7a2be512bf3)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=8b2fd4b920d0bb1e6d9c839a1da0a6b5f5c1b118)<br/>[CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=2021-22600)|
|2021-01-06|2021-01-06|[Patch](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=722d94847de29310e8aa03fcbdb41fc92c521756)|[checksum](#60f8c7ce4b322a0c92aaf96f9f2363508ba70db2cc96483dbf989d822323998e)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=53c05996968fc87df17de205b461f4f96d5b5907)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=852ddf3aee4937a946abb6b1331c4336122981b9)<br/>[CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-0185)|
|2021-01-14|2021-01-18|[Patch](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=722d94847de29310e8aa03fcbdb41fc92c521756)|[checksum](#76a3b66a6272dccbb7c6b6d89683845b4e3c47cb13bda4e0e8095e9b9adebc38)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=53c05996968fc87df17de205b461f4f96d5b5907)<br/>[Syzkaller](https://syzkaller.appspot.com/bug?id=852ddf3aee4937a946abb6b1331c4336122981b9)<br/>[CVE](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-0185)|
|2022-01-22|?|?|[checksum](#5859a79450bdebdaa4a9f41c187ab5cafd2666f102d3cdb77382b0bd5234ae86)|
|2022-01-24|?|?|[checksum](#f3c5659e93474c9434a7afb8120f139d78930fc2)|


In case of questions or suggestions, you can reach us in [#kctf](https://discord.gg/V8UqnZ6JBG).
