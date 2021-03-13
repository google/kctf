# Local Testing Walkthrough

In this walkthrough, you will learn how to use the kCTF infrastructure.

In detail, this means:

1. [First-time setup](#first-time-setup): Setting the right `umask`, installing dependencies, enabling user namespaces.
2. [Using kCTF](#using-kctf): Downloading the SDK, creating a local cluster and creating tasks.
3. [Debugging](#debug-failures) any errors you encountered. 

## First-time setup

Following this walkthrough requires a local Linux machine capable of running Docker.

### Set the right umask

Since nsjail will run commands as another user, the challenge files need to be readable by all users. To enable this, set the following umask:
```
umask a+rx
```

This should also be setup in our `.bashrc` and `.bash_profile` to ensure you never forget to set it.

```
echo umask a+rx | tee -a ~/.bashrc ~/.bash_profile
```

If you forget to set the right umask, the SDK will warn you.

### Install dependencies
Most people should have `wget`, `curl` and `xxd` installed already, but if you are running on a fresh Debian run:
```
sudo apt install xxd wget curl netcat
```

### Install Docker
```
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER && newgrp docker
```

### Enable user namespaces
Some Linux distributions don't have user namespaces by default, to enable them in Debian or Ubuntu execute:
```
echo 'kernel.unprivileged_userns_clone=1' > /etc/sysctl.d/00-local-userns.conf
service procps restart
```

## Using kCTF
### Downloading and activating kCTF
```
mkdir ctf-directory && cd ctf-directory
curl //kctf.dev/sdk_1_0_0 | tar xzf
source kctf/activate
```

### Create a sample local cluster
```
kctf cluster create local-cluster --start --type kind
```

### Create basic demo challenge
```
kctf chal create chal-sample && cd chal-sample
kctf chal start
```

### Connect to the challenge
This will take a bit longer the first time it is run, as it has to build a chroot.
```
kctf chal debug port-forward &
nc 127.0.0.1 [external_port]
```

When `kctf chal debug port-forward` runs, you will see `0.0.0.0:[external_port]->1337/tcp` in the terminal.

If all went well, you should be able to connect. If you don't, there might be some issues with your system (support for nsjail, Docker etc.). See the [debugging instructions](#errors-with-docker) further down on this page.

## Debug failures

The instructions below can help you resolve errors in the setup.

### Permission denied in nsjail

If you see an error in the Docker logs that says that it can't access /config/nsjail.cfg:
```
[W][2020-02-13T15:14:36+0000][1] bool config::parseFile(nsjconf_t*, const char*)():300 Couldn't open config file '/config/nsjail.cfg': Permission denied
```

This probably means that you didn't [set your umask](#set-the-right-umask) correctly.

### CLONE error in nsjail

If you have the following type of error:
```
[E][2020-01-31T20:16:39+0000][1] bool subproc::runChild(nsjconf_t*, int, int, int)():459 nsjail tried to use the CLONE_NEWCGROUP clone flag, which is supported under kernel versions >= 4.6 only. Try disabling this flag: Operation not permitted
[E][2020-01-31T20:16:39+0000][1] bool subproc::runChild(nsjconf_t*, int, int, int)():464 clone(flags=CLONE_NEWNS|CLONE_NEWCGROUP|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWUSER|CLONE_NEWPID|CLONE_NEWNET|SIGCHLD) failed. You probably need root privileges if your system doesn't support CLONE_NEWUSER. Alternatively, you might want to recompile your kernel with support for namespaces or check the current value of the kernel.unprivileged_userns_clone sysctl: Operation not permitted
```
This probably means that unprivileged user namespaces are not enabled. You can fix this by running:
```
(echo 1 | sudo tee /proc/sys/kernel/unprivileged_userns_clone) || (echo 'kernel.unprivileged_userns_clone=1' | sudo tee /etc/sysctl.d/00-local-userns.conf) 2>&1
sudo service procps restart
```
Then try connecting through netcat again.

### Errors inside the challenge
If you see errors like the following:
```
bash: cannot set terminal process group (-1): Inappropriate ioctl for device
bash: no job control in this shell
```

That's normal, just ignore them. You should still get a shell afterwards.
