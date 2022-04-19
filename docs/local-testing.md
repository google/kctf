# Local Testing Walkthrough

In this walkthrough, you will learn how to use the kCTF infrastructure.

In detail, this means:

1. [First-time setup](#first-time-setup): Setting the right `umask`, installing dependencies, and enabling user namespaces.
2. [Using kCTF](#using-kctf): Downloading the SDK, creating a local cluster, and creating tasks.
3. [Troubleshooting](#Troubleshooting): Basic `kctf chal debug` commands and `kubectl` usage.

## First-time setup

Following this walkthrough requires a local Linux machine capable of running Docker.

### Set the right umask

Since nsjail will run commands as another user, the challenge files need to be readable by all users. To enable this, set the following umask:
```bash
umask a+rx
```

If you forget to set the right umask, the SDK will warn you.

### Install dependencies
Most people should have `wget`, `curl`, and `xxd` installed already, but if you are running on a fresh Debian, run this command:
```bash
sudo apt install xxd wget curl netcat
```

### Install Docker
If you have not installed Docker, you should do so by following the [official instructions](https://docs.docker.com/engine/install/).

At the time of writing, one of the supported methods to install Docker is by running the following commands:
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER && newgrp docker
```

If you already have Docker installed, the script will show a warning.

### Enable user namespaces
Some Linux distributions don't have user namespaces by default. To enable them in Debian or Ubuntu, execute:
```bash
echo 'kernel.unprivileged_userns_clone=1' | sudo tee -a /etc/sysctl.d/00-local-userns.conf
sudo service procps restart
```

Note that this has some security implications due to an increased kernel attack surface, which is why Debian and Ubuntu don't enable them by default.

## Using kCTF
### Downloading and activating kCTF
kCTF has an SDK that requires you to install kCTF in a directory. All challenges should be under the directory where the kCTF SDK is installed.
```bash
mkdir ctf-directory && cd ctf-directory
curl -sSL https://kctf.dev/sdk | tar xz
source kctf/activate
```

After you've done that, you should see kCTF is enabled in the prompt:
```
evn@evn:~/ctf-directory$ kCTF[ctf=ctf-directory] > 
```

To exit from the environment, run `deactivate`.

### Create a sample local cluster
To run local challenges, you need to create a local Kubernetes cluster. Do so by running:
```bash
kctf cluster create local-cluster --start --type kind
```

In the prompt, you should be able to see that it worked correctly (notice `config=local-cluster`):
```
evn@evn:~/ctf-directory$ kCTF[ctf=ctf-directory,config=local-cluster] > 
```

### Create basic demo challenge
To create a challenge from a skeleton, you can run the following command:
```bash
kctf chal create chal-sample && cd chal-sample
```

This creates a sample `pwn` challenge, but you can create `web` or `xss-bot` templates as well with the `--template` parameter.

You should then notice that the prompt detected that you are inside of a challenge directory (notice `chal=chal-sample`):
```
evn@evn:~/ctf-directory/chal-sample$ kCTF[ctf=ctf-directory,config=local-cluster,chal=chal-sample] > 
```

The pwn template comes with a Makefile to build the challenge binary. This is
recommended if you want to hand out the binary as an attachment to players,
e.g. since the layout might matter for ROP gadgets. If the layout doesn't
matter, you could also build it in an intermediate container as part of your
Dockerfile.

To start the challenge, run:
```bash
make -C challenge && kctf chal start
```

Once the challenge is built and deployed, you will see `challenge.kctf.dev/chal-sample created`.

### Connect to the challenge
To connect to the challenge, run the following command:
```bash
kctf chal debug port-forward &
```

After connecting, you will see `Forwarding from 127.0.0.1:[LOCAL_PORT] -> 1337` in the terminal. Connect to the LOCAL_PORT:
```bash
nc 127.0.0.1 [LOCAL_PORT]
```

If all went well, you should be able to connect and see:
```
== proof-of-work: disabled ==
CTF{TestFlag}
```

## Troubleshooting

### Modify the challenge
To test what would happen if the challenge broke, we can modify the task, and see what happens.

Run the following command to replace `cat /flag` with `echo /flag` in `challenge/chal.c`:
```bash
sed -i s/cat/echo/ challenge/chal.c && make -C challenge
```

If you try to connect again, you will notice that the old version is still running (you will still see `CTF{TestFlag}`).

This is because the new update is made as a rollout, and the old challenge is only stopped once the new one is ready.

**If you see an old version of the challenge running, it means that the deployment didn't work.**

### Check challenge status
To see the status of the challenge deployment, you can run:
```bash
kctf chal status
```

The preceding command should return something like this (notice it says "unhealthy" and that the most recent POD is READY=`1/2`):
```
= CHALLENGE RESOURCE =

NAME          HEALTH      STATUS    DEPLOYED   PUBLIC
chal-sample   unhealthy   Running   true       false

= INSTANCES / PODs =

Challenge execution status
This shows you how many instances of the challenges are running.

NAME                           READY   STATUS    RESTARTS   AGE     IP            NODE                         NOMINATED NODE   READINESS GATES
chal-sample-66cb778c45-stjkq   2/2     Running   0          5m49s   10.244.0.10   kctf-cluster-control-plane   <none>           <none>
chal-sample-6c86956d78-5md9s   1/2     Running   0          15s     10.244.0.11   kctf-cluster-control-plane   <none>           <none>


= DEPLOYMENTS =

Challenge deployment status
This shows you if the challenge was deployed to the cluster.

NAME          READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS              IMAGES                                                                                                                                                              SELECTOR
chal-sample   1/1     1            1           5m49s   challenge,healthcheck   kind/challenge:6eba2,kind/healthcheck:dcf4   app=chal-sample

= EXTERNAL SERVICES =

Challenge external status
This shows you if the challenge is exposed externally.

SERVICES:
NAME          TYPE       EXTERNAL-IP   PORT   DNS
chal-sample   NodePort   <none>        1337   <none>

Ingresses:
No resources found in default namespace.

```

### Reading logs
To troubleshoot unhealthy challenges, we need to read the logs of the failing healthcheck. We need to do that using `kubectl`.
```bash
kubectl logs chal-sample-6c86956d78-5md9s -c healthcheck
```

Replace `chal-sample-6c86956d78-5md9s` with the name of the pod that says READY=`1/2`. You will see something like this:
```
[Sat Mar 13 12:19:58 UTC 2021] b'== proof-of-work: '
Traceback (most recent call last):
  File "/home/user/healthcheck.py", line 33, in <module>
    print(r.recvuntil(b'CTF{'))
  File "/usr/local/lib/python3.8/dist-packages/pwnlib/tubes/tube.py", line 310, in recvuntil
    res = self.recv(timeout=self.timeout)
  File "/usr/local/lib/python3.8/dist-packages/pwnlib/tubes/tube.py", line 82, in recv
    return self._recv(numb, timeout) or b''
  File "/usr/local/lib/python3.8/dist-packages/pwnlib/tubes/tube.py", line 160, in _recv
    if not self.buffer and not self._fillbuffer(timeout):
  File "/usr/local/lib/python3.8/dist-packages/pwnlib/tubes/tube.py", line 131, in _fillbuffer
    data = self.recv_raw(self.buffer.get_fill_size())
  File "/usr/local/lib/python3.8/dist-packages/pwnlib/tubes/sock.py", line 56, in recv_raw
    raise EOFError
EOFError
1 err
```

Here you can see that the healthcheck is failing to read `CTF{` from the challenge (as we changed `cat /flag` to `echo /flag`).

In order to facilitate development, you can disable the healthcheck by running the following command:
```bash
sed -i s/enabled:\ true/enabled:\ false/ challenge.yaml
```

Then run `kctf chal start` once again. If you now try to connect again, you will see:
```
== proof-of-work: disabled ==
/flag
```
