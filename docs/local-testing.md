# Local Testing Walkthrough

In this walkthrough, you will learn how to use the kCTF infrastructure.

In detail, this means:

1. Setting the right `umask`.
2. Deciding on your test environment ([Docker](#testing-with-docker-only) or [Kubernetes](#testing-with-kubernetes)) and performing the related steps.
3. [Debugging](#debug-failures) any errors you encountered. 

## Set the right umask

Since nsjail will run commands as another user, the challenge files need to be readable by all users. To enable this, set the following umask:
```
umask a+rx
```

## Testing with Docker only

Following this walkthrough requires a local Linux machine capable of running Docker.

This is the fastest way to get started with developing challenges. Alternatively, you can [test with a local Kubernetes cluster](#testing-with-kubernetes) for a more precise emulation of the production environment.

### Install Docker
```
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER && newgrp docker
```

### Download kCTF
```
git clone https://github.com/google/kctf.git
PATH=$PATH:$PWD/kctf/bin
```

### Create basic demo challenge
```
CHALDIR=$(mktemp -d)
kctf-setup-chal-dir $CHALDIR
kctf-chal-create test-1
```

### Test connecting to the challenge
This will take a bit longer the first time it is run, as it has to build a chroot.
```
cd $CHALDIR/test-1
make test-docker
sudo apt-get install -y netcat
nc 127.0.0.1 [external_port]
```

When `make test-docker` runs, you will see `0.0.0.0:[external_port]->1337/tcp` in the terminal.

If all went well, you should have a shell inside an nsjail bash (use Ctrl+C to exit):
```
== kCTF challenge skeleton ==
Flag: CTF{TestFlag}
=========================
```

If you don't, there might be some issues with your system (support for nsjail, Docker etc.). See the [debugging instructions](#errors-with-docker) further down on this page.

## Testing with Kubernetes

To test challenges in the same environment as production, and to test challenges that require changes to the Kubernetes configuration, you can test with a local Kubernetes cluster. This is the recommended method, although it takes a bit longer to set up, and is not available in Google Cloud Shell.

### Install kubectl 1.17+

Download the latest version:
```
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
```

Make it executable and move it to your PATH:
```
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

### Install a local Kubernetes cluster

There are several options for installing a local Kubernetes cluster:

1. **KIND** – On Linux, use the following command:
    ```
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$(uname)-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
    ```
    MacOS users can follow the same approach, but might need to adjust the path. Windows users can [follow these instructions](https://kind.sigs.k8s.io/docs/user/quick-start/).
1. **Docker for Desktop** – Windows users can try the built-in Kubernetes cluster. See the [WSL1 instructions](#wsl1).

### Running the challenge in Kubernetes
Once you run the command to create the cluster (`kind create cluster`), create a configuration for one of the challenge samples (e.g. the one located in `kctf/samples/apache-php` folder) using the following command:
```
kctf-config-create --chal-dir [directory_of_the_challenge] --project test-kind-kctf
```

Then, run the challenge sample inside the cluster calling the following command inside the challenge folder (in the challenge suggested, it would be inside the `samples/apache-php` folder) using:
```
make test-kind
```

This deploys the challenge and healthcheck to the local Kubernetes cluster.

### Connect to the challenge

Use the following command to connect to the challenge:
```
kubectl port-forward --namespace=[name_of_the_challenge] deployment/chal :1337 &
```
Example: If you ran the challenge sample in the `samples/apache-php` folder, specify `apache-php` as the namespace. After you run the `port-forward` command, the command line displays from which external port the challenge is being forwarded. Based on that information, you can access `127.0.0.1:[external_port]` from your browser and should see a page similar to this one:
![Apache PHP sample page](https://raw.githubusercontent.com/google/kctf/master/docs/images/php_sample.png)

## Debug failures

The instructions below can help you resolve errors in the setup.

### Errors with Docker

#### Errors installing Docker

If you get this error installing Docker on Ubuntu:
```
E: Package 'docker-ce' has no installation candidate
```
Try to install Ubuntu's version of Docker:
```
sudo apt-get install -y docker.io
```

#### Permission denied on /var/run/docker.sock

If you get an error like this:
```
Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Post http://%2Fvar%2Frun%2Fdocker.sock/v1.40/build?buildargs=%7B%7D&cachefrom=%5B%5D&cgroupparent=&cpuperiod=0&cpuquota=0&cpusetcpus=&cpusetmems=&cpushares=0&dockerfile=Dockerfile&labels=%7B%7D&memory=0&memswap=0&networkmode=default&rm=1&session=&shmsize=0&t=kctf-nsjail&target=&ulimits=null&version=1: dial unix /var/run/docker.sock: connect: permission denied
```

Your user doesn't have permission to run Docker. To fix this, run:
```
sudo usermod -aG docker $USER && newgrp docker
```

#### Docker daemon error
In some cases, Docker leaves the socket on the FS, but it's no longer running. The related error looks like this:
```
Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
```

To fix this, run:
```
sudo service docker restart
```

#### IPv4 forwarding error
In some cases, Docker can fail to run because of network errors. Check if you have this error:
```
 ---> [Warning] IPv4 forwarding is disabled. Networking will not work.
```
If so, then you have to enable ipv4 forwarding, by running:
```
echo net.ipv4.ip_forward=1 | sudo tee -a /etc/sysctl.conf
```
And since this probably made Docker cache an invalid `apt-get update`, you also have to run `docker system prune -a` before running the `kctf-chal-test-docker` command again.

### Errors connecting to the challenge
If you can't connect, enter:
```
docker ps -a
```
In the table, look for the container which was the last to run, and then run:
```
docker logs CONTAINER_NAME
```
replacing CONTAINER_NAME with the name of the container which ran last.

This outputs the logs from the last time the container ran.

#### Permission denied in nsjail

If you see an error in the Docker logs that says that it can't access /config/nsjail.cfg:
```
[W][2020-02-13T15:14:36+0000][1] bool config::parseFile(nsjconf_t*, const char*)():300 Couldn't open config file '/config/nsjail.cfg': Permission denied
```

This probably means that you didn't [set your umask](#set-the-right-umask) correctly.

#### CLONE error in nsjail

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

### Errors in Windows Subsystem for Linux

#### WSL1

The only **supported** method (see below for unsupported options) for testing on WSL1 is through a local Kubernetes cluster, and this is only possible for users with a Pro/Edu/Enterprise version of Windows 10. This is because [Docker for Windows](https://docs.docker.com/docker-for-windows/) requires [Hyper-V support](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/about/).

You must [download Docker for Windows](https://download.docker.com/win/stable/Docker%20Desktop%20Installer.exe) (1GB in size). It already includes Kubernetes, but the cluster has to be manually started in the UI. Go to *Settings > Kubernetes* and click on **Enable Kubernetes**.

Open WSL 1 and enter:
```
echo -e '#!/bin/bash\n$(basename $0).exe $@' | tee  $HOME/.local/bin/docker $HOME/.local/bin/kubectl
chmod +x $HOME/.local/bin/docker $HOME/.local/bin/kubectl
```

Make sure you are running the Windows version of Docker and kubectl:
<!-- {% raw  %} -->
```
docker version --format {{.Client.Os}} 
kubectl version --client -o yaml | grep platform
```
<!-- {% endraw  %} -->
If the commands say Linux, you must uninstall the local Linux versions, and make sure `~/.local/bin` is in your PATH.

To run on Docker for Windows, use the command:
```
make test-d4w
```

That launches the Docker image in a Kubernetes cluster.

##### WSL1 unsupported flow

You can try an **unsupported** option for running Docker: [Follow this guide](https://medium.com/faun/docker-running-seamlessly-in-windows-subsystem-linux-6ef8412377aa).

This is the only option available for WSL1 users on Windows 10 Home (other than running a VirtualBox VM).

#### WSL2

WSL2 should work out of the box with KIND, but WSL2 requires the user to be enrolled in Windows Insider, which has extra data collection by Microsoft and more Windows updates. The setup of WSL2 takes a while and requires multiple restarts.

1. Register in [Windows Insider](https://insider.windows.com/en-us/register/)
1. [Upgrade WSL to version 2](https://docs.microsoft.com/en-us/windows/wsl/wsl2-install)
1. [Enable the WSL2 Docker engine](https://docs.docker.com/docker-for-windows/wsl-tech-preview/)

To run on Docker for Windows, use the command:
```
make test-d4w
```

Alternatively, the instructions for running KIND also work for WSL2 users.

### Errors inside the challenge
If you see errors like the following:
```
bash: cannot set terminal process group (-1): Inappropriate ioctl for device
bash: no job control in this shell
```

That's normal, just ignore them. You should still get a shell afterwards.
