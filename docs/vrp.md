# kCTF VRP Setup

We invite you to demonstrate an exploit against our kCTF demo cluster based on the challenges presented on this page. Successful demonstrations are eligible for rewards between 5,000 to 10,000 USD as defined in https://g.co/vrp.

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


It will ask you to solve a proof-of-work and then gives you access to a bash running in a setup similar to the [kCTF pwn template challenge](https://github.com/google/kctf/tree/beta/dist/challenge-templates/pwn). The only difference is that the flag is not accessible inside of the nsjail sandbox and you will need to break out of the chroot in order to read it.


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
  2. Once you have a working exploit, test it on the [lab environment](#kctf-challenge). If you have troubles let us know [here](https://github.com/google/kctf/issues) and we'll help you figure out any problems.
  3. Once you get the flag, send it [here](https://docs.google.com/forms/d/e/1FAIpQLSeQf6aWmIIjtG4sbEKfgOBK0KL3zzeHCrsgA1EcPr-xsFAk7w/viewform). You won't share any technical details about the vulnerability, you will just record the fact you found something (as we only reward the first person that writes an exploit for a given bug, we use it to resolve the timing in case of an exploit collision).
  4. Once the vulnerability is fixed, please contact us at [g.co/vulnz](https://g.co/vulnz) and include:
     1. The patch that fixed the vulnerability.
     2. The exploit you used to trigger it.

### Notes

We want to encourage the community to help research vulnerabilities such as those [found by Syzkaller](https://syzkaller.appspot.com/upstream#open), but which are still unfixed since they have not been shown to be exploitable. As such:



*   The person that develops the exploit and receives the reward might not be the same as the person that discovered or patched the vulnerability.
*   It's ok to use 1-day exploits against the lab environment using publicly known vulnerabilities that exploit the patch gap between the time when a patch is announced and the lab environment is updated, however we will only issue a single reward per vulnerability.


In case of questions or suggestions, you can reach us at kctf@google.com
