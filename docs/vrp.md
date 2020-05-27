# kCTF VRP Setup

Eligible exploits should be demonstrated against our kCTF demo cluster and on this page we will describe the setup.

[kCTF](https://github.com/google/kctf) is open source infrastructure for CTF competitions. You can find details on how it works in the [kCTF documentation](https://google.github.io/kctf/introduction.html) but in short, it’s running on a hardened kubernetes cluster with the following security features:

*   The OS and kubernetes versions are upgraded automatically.
*   The nodes are running Container-Optimized OS.
*   The pod egress network access is restricted to public IPs only.
*   [Workload Identity](https://cloud.google.com/blog/products/containers-kubernetes/introducing-workload-identity-better-authentication-for-your-gke-applications) restricts access to service accounts and the metadata server in addition to the network policies.
*   Every connection to a challenge spawns a separate [nsjail](https://github.com/google/nsjail) sandbox to isolate players from each other.

At present, we’re interested in two attack scenarios against this infrastructure:

1. Breaking out of the nsjail sandbox as it would allow solving challenges in unintended ways.
2. Breaking the isolation that kubernetes provides and accessing the flags from other challenges.

For that we set up two kCTF challenges with secret flags: “kctf” and “full-chain”. You can demonstrate a working exploit by leaking the flags of either of these.

![drawing showing the location of the flags](./images/flag-locations.png)


## kctf challenge

The “kctf” challenge is the only entry point to the cluster. You can connect to it via


```
nc kctf.vrp.ctfcompetition.com 1
```


It will ask you to solve a proof-of-work and then give you access to a bash running in a setup similar to the [kCTF bash example challenge](https://github.com/google/kctf/tree/master/samples/bash). The only difference is that the flag is not accessible inside of the nsjail sandbox and you will need to break out of the chroot in order to read it.


## full-chain challenge

The “full-chain” challenge is a challenge that just runs a `while sleep` loop and doesn’t have any exposed ports. In order to get access to the flag, you will need to break out of the “kctf” challenge and break the pod isolation of the cluster.


## Flags

The flags are stored in kubernetes [secrets](https://kubernetes.io/docs/concepts/configuration/secret/) and mounted to the filesystem of the two challenges at “/flag/flag”. They are of the format:


```
KCTF{$CHAL_NAME-$TIMESTAMP:$MAC}
```


As you can see, the flags include a timestamp and are rotated frequently. You can send us the flag [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform) before you are ready to disclose the exploit as we can use it to resolve the timing in the case of an exploit collision (we will reward whoever was the first to obtain and record the submission of a flag), otherwise, once the vulnerability is fixed, please contact us at [g.co/vulnz](g.co/vulnz).


### Notes

We want to encourage the community to help research vulnerabilities such as those [found by Syzkaller](https://syzkaller.appspot.com/upstream), but which are still unfixed since they have not been shown to be exploitable. As such:



*   The person that develops the exploit and receives the reward might not be the same as the person that discovered or patched the vulnerability.
*   It's ok to use 1-day exploits against the lab environment using publicly known vulnerabilities that exploit the patch gap between the time when a patch is announced and the lab environment is updated, however we will only issue a single reward per vulnerability.
