# Security Policy

To report a vulnerability in this repository, or in Google Cloud please contact the Google Security Team at https://g.co/vulnz

To report a vulnerability in a dependency, please contact the upstream maintainer directly.

 - [Kubernetes](https://kubernetes.io/docs/reference/issues-security/security/#report-a-vulnerability)
 - [Docker](https://github.com/moby/moby/blob/master/CONTRIBUTING.md#reporting-security-issues)
 - [Ubuntu](https://wiki.ubuntu.com/SecurityTeam/FAQ?_ga=2.254550412.542177495.1583355140-2013298171.1583355140#Contact)
 - [Linux kernel](https://www.kernel.org/doc/html/v4.11/admin-guide/security-bugs.html)

## Security Threat Model

kCTF is a Kubernetes-based infrastructure in which authors are able to expose CTF challenges with security vulnerabilities like SQL injection or even with code execution. kCTF provides tools so that it should be impossible for an attacker to disturb the challenge, sabotage it or gain an advantage over other players. In order to accomplish this, the infrastructure provides tools to sandbox and isolate every instance of a challenge from other users.

### Assets

The assets that kCTF manages are:
 - **The flags**: If an attacker gains access to the flags, this would undermine the whole competition. Recovery from this state, if detected, could be as simple as rotating the challenges.
 - **The challenges**: If an attacker gains access to the challenges source code (when unintended), this could unfairly benefit the attackers. Recovery from this state could be impossible for challenges on which the source code is meant to be secret.
 - **The exploits**: In order to ensure that challenges are working, the infrastructure asks authors to provide an exploit to serve as a healthcheck. Recovery from this state is impossible, as this leaks the solutions to the tasks.
 - **Players solutions**: Similar to exploits, if an attacker is able to read the solutions of other players, this could undermine the competition. Recovery from this state is also impossible.
 - **Availability**: If the a task or the infrastructure suffer a Denial of Service attack, this would have as a consequence losing the competition. Recovery from this state might be too expensive.

![threat model visualization](https://raw.githubusercontent.com/google/kctf/master/docs/images/threat-model-graph.png)

### Risks

#### Players -> Infrastructure

The players interact with the challenges through a Load Balancer configured in Google Compute Engine (GCE), which routes traffic to different virtual machines. Players' internet connection could be compromised (eg, with a man-in-the-middle attack), or the players themselves could be compromised.

As such, the risks on this edge are:
 - Players internet connection might be compromised
 - Players could be compromised

Ways to limit these risks are:
 - Players could connect to the tasks through HTTPS. There are no plans to do this.

The compromise of the players themselves is out of scope of kCTF's threat model.

#### Infrastructure -> Players solutions

The environment from every player should be isolated from other players. kCTF usually does this by creating a new Linux user namespace for every TCP connection. Some challenges, however, require persistent storage. Challenges that require persistent storage in between requests has the consequence of relaxing these guarantees. This means that in those cases challenges have to be carefully designed so that there is isolation in between the sandboxed environment, and the environment managing the shared storage. This is because as players solve the challenges, they might leave some traces in the infrastructure that other players could find and reuse solve the challenges faster.

As such, the risks on this edge are:
 - Challenges might leave other players solutions on the server
 - Isolation between users might not be configured properly

Ways to limit these risks are:
 - kCTF provides a per-TCP connection isolation mechanism by default.
 - kCTF could provide a mechanism to isolate tasks per-team, rather than per-TCP connection. There are no plans to do this.
 - kCTF could detect dangerous misconfigurations and warn authors. There are no plans to do this.

#### Infrastructure -> Availability

The infrastructure runs on Kubernetes, which supports automatic scaling and provisioning of resources. It also supports "proof of work" setups that can limit the impact of a malicious targeted attack. However, either through misconfiguration, or lack of monitoring, an attacker could exhaust the resources of a competition.

As such, the risks on this edge are:
 - Challenges might consume too many resources.
 - Authors might respond too slowly to DoS attacks.

Ways to limit these risks are:
 - kCTF automatically scales and limits the number of maximum nodes in a cluster.
 - kCTF could automatically enable the proof-of-work when an attack is detected. There are no plans to do this.

#### Infrastructure -> Kubernetes

The infrastructure itself runs on top of the kubernetes control plane. As such, the compromise of the system from a challenge is possible (eg, through a vulnerability in the Linux kernel, or Kubernetes itself).

As such, the risks on this edge are:
 - Linux kernel or Kubernetes might have a vulnerability.
 
Ways to limit these risks are:
 - kCTF automatically keeps its kernel up-to-date.
 - kCTF could limit (through seccomp or similar) the kernel attack surface. There are no plans to do this.

#### Authors -> Kubernetes

The authentication between authors and Kubernetes is managed by Google Kubernetes Engine (GKE). Any vulnerability in GKE that allows unauthorized users to access GKE could gain access to all challenges code, the flags and exploits. In addition, any author that is compromised can undermine, similarly the competition.

As such, the risks on this edge are:
 - author could be compromised
 - kubernetes or GKE might have a vulnerability

Ways to limit these risks are:
 - kCTF keeps all clusters always up-to-date, and follows hardening and configuration best practices.
 - kCTF could use least privilege principles, and limit access of authors to just their own tasks. Work to do this is tracked in [issue #29](https://github.com/google/kctf/issues/29).

#### Kubernetes -> Challenges

The challenges are stored as docker images in a private docker registry on top of Google Container Registry (GCR). Any vulnerabilities in GCR or Google Cloud Storage (GCS) could result in the compromise of the challenges.

As such, the risks on this edge are:
 - GCS or GCR might have vulnerabilities
 - Project might be misconfigured

Ways to limit these risks are:
 - kCTF could limits the information in docker images to the minimum possible. There are no plans to do this.
 - kCTF could check configuration to detect accidental misconfiguration. There are no plans to do this.

#### Kubernetes -> Flags

The flags are stored in the challenge images. The same considerations from the previous section apply.

#### Kubernetes -> Exploits

The exploits are stored in the challenge images. The same considerations from the previous section apply.

### Attackers
The usual type of attackers to a system of this form mainly fall into two categories:
 - Attention-seeking malicious disruptive attackers
 - Secretive cheating-driven attackers

#### Cost of Vulnerability Research
Finding and exploiting a vulnerability in kCTF provides different benefits to the attackers, usually reputational. The barrier of entry to perform an attack usually is the cost to find the vulnerability. In order to increase the cost of these attacks, encouraging broad research by the well-intentioned security community on the same areas, could result in a lower likelihood of an attacker paying the cost of a successful attack.

#### Cost of Operations
Conducting an attack usually consists of a cycle of trial and error. Although kCTF is open-source (which reduces this to some degree), in practice the cost of being stealthy is directly correlated to the amount of preparation required. If the kCTF infrastructure provides thorough monitoring and alerting, this can increase the cost of preparation for the attackers. Work to do this is tracked in [Issue #30](https://github.com/google/kctf/issues/30).

#### Return of Investment
Disrupting a competition done for the benefit of the security community is not immediately obvious, unless the competition itself, or the organizers, have managed to incentivise some of the members of the community to go through the effort to perform such an attack. As such, running a fair and high quality competition is an easy way for organizers to prevent alienating and making skilled players convert from users to attackers.

### Conclusions

kCTF attempts to address and balance several security concerns with usability and complexity. Understanding and improving the security of kCTF will be a constant process that will be done over time with the feedback of the community as different problems and solutions are discovered. This document will be kept up-to-date as new information arises.
