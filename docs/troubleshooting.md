# Troubleshooting

Users are reporting a challenge is down, the author is offline, and you don't know how long the challenge has been broken. The challenge had no healthcheck configured, and there's no documentation. Someone forgot to test the task. A nightmare come true.

This guide will show you how to troubleshoot a broken challenge, assuming you don't know how the challenge works. The guide is divided into three parts that take you through troubleshooting a task in different environments:
* [In Docker](#troubleshooting-with-docker) – Testing locally with `kctf chal debug docker` is the easiest and fastest way to troubleshoot challenges, and you don't need to know much apart from some basic Docker commands presented in the following section.
* [In a local (Kubernetes) cluster](#troubleshooting-with-kubernetes) – Troubleshooting with Kubernetes requires a few extra setup steps. However, this is needed only on rare occasions, as it is only relevant if the author made changes to the Kubernetes setup, or if there's a bug in kCTF.
* [Remotely](#troubleshooting-remotely) – Troubleshooting remotely is trivial, although it runs the risk of the user leaving the remote state in an inconsistent state, however it's a good last-resort.

Note: The commands in this guide use the `kctf-chal-troubleshooting` placeholder as the name of the broken challenge, replace this placeholder with the actual name of your challenge.

Remember to run `source kctf/activate` on your CTF directory before running any of the following commands.

## Troubleshooting with Docker

### Building and running Docker

A good place to start is to check if the Docker image works. It can happen that an author makes a small change to a Dockerfile which then breaks the task. A challenge that doesn't even start usually indicates that something is broken.

To build the Docker image, run:
```
kctf chal debug docker
```

This will output any errors when *building* the image and something like this towards the end:
```
CONTAINER ID        IMAGE                       COMMAND                  CREATED             STATUS                  PORTS                     NAMES
81d06c71f6b5        kctf-chal-troubleshooting   "/bin/sh -c '/usr/bi…"   1 second ago        Up Less than a second   0.0.0.0:32780->1337/tcp   unruffled_payne
```

Notice that there's a PORTS column that tells you that port 32780 in your local machine is mapped to port 1337 inside the container. The challenge should work if you connect to the local port (exchange "32780" with your local port in the following command):

```
nc localhost 32780
```

If the challenge is very broken, it still won't work. If you enter anything, nothing will happen. Another possible way in which the challenge could be broken, is if there was an error running the command, in which case connecting to the challenge will also not work (it would just exit).

You can actually check if the challenge is still running a few seconds after starting it, by running:
```
docker ps -f ancestor=kctf-chal-troubleshooting
```

If you don't see the challenge running anymore, that's a good indication that the task failed to even start. If the challenge is still running, but connecting to it fails, then the challenge started properly but there's some runtime error.

Either way, the next step would be to read the Docker logs.

### Looking at the logs

If the challenge builds, but running it doesn't work, the best thing to try is to read the Docker logs. To get the Docker logs, run:
```
docker logs $(docker ps -q -f ancestor=kctf-chal-troubleshooting)
```

If the challenge is a pwnable, you might see an error formatted like this:
```
[E][2020-02-02T20:20:02+0200][1] void subproc::subprocNewProc(nsjconf_t*, int, int, int, int)():204 execve('/bin/sh') failed: No such file or directory
```

That's an nsjail error. In this case, it is complaining that `/bin/sh` failed and returned the error *No such file or directory*.

If your challenge is based on the `apache-php` example, you will see Apache access/error logs instead, but either way, to continue to debug you need to find out what is going on inside the challenge.

### Shell into the Docker image

To continue debugging, you can get a "shell" into the container with `docker exec`. To do so run:
```
docker exec -it $(docker ps -q -f ancestor=kctf-chal-troubleshooting) bash
```

Once inside, you can inspect the environment in which the challenge is running. For example, you can list the current processes:
```
ps aux
```

This way you can find out what is currently running on the task, but you could also find this out by reading the Dockerfile and checking the `CMD` that is configured there.

Here you can also inspect other logs (if any) as well as check the permissions on the filesystem. The next step in debugging would be to run the command on the shell directly and see if you can get new information.

The next troubleshooting step is to find out where nsjail is being called. For pwnables, nsjail will be running directly as the listening service, and for web tasks, it would run as a CGI script when PHP files are executed.

### Debugging nsjail

The configuration for nsjail usually lives in /config. You will want to run nsjail again with the same command that Docker attempted, but for testing. Nsjail supports overriding the configuration via command line arguments added after the `--config` flag. If nsjail is configured to work in "listening" mode (e.g. listening on a port), then you can override that to run in `run_once` mode by adding the flag `-Mo`:
```
/usr/bin/nsjail --config /config/nsjail.cfg -Mo
```

This should trigger the same error as the one found in `docker logs` from the previous step, however, now you can enable more verbose options if you run:
```
/usr/bin/nsjail --config /config/nsjail.cfg -Mo -v
```

This should output more errors, and these errors should provide more details about what went wrong.

## Troubleshooting with Kubernetes

If everything works in Docker, the problem might be higher up (in Kubernetes). The first step to debug this would be to check if the challenge works in a local cluster. Follow the instructions [here](local-testing.md#running-the-challenge-in-kubernetes) for getting the task running in [KIND](https://github.com/kubernetes-sigs/kind).

### Basic commands

Once the local cluster is running, you can follow similar steps as above for debugging.

First, change to the namespace of the challenge by running:
```
kubectl config set-context --current --namespace=kctf-chal-troubleshooting
```

To retrieve the status of the challenge, run:
```
kubectl get deployment/chal
```

To read execution logs, run:
```
kubectl logs deployment/chal -c challenge
```

To obtain a shell into a pod, run:
```
kubectl exec -it deployment/chal -c challenge
```

However, there are a few additional commands for debugging Kubernetes-specific errors.

### Inspecting a Kubernetes Deployment

A basic understanding of Kubernetes (as described in [kCTF in 8 minutes](introduction.md)) is a useful prerequisite, but this guide should be intuitive enough to roughly understand what is happening even without knowledge of Kubernetes.

The most common way to troubleshoot will be using the `kubectl describe` command. This command tells you everything Kubernetes knows about a challenge. You should start by describing the "deployment" by running:
```
kubectl describe deployment/chal
```

The most interesting parts of this command are the:
 - *Status* and *Reason*
 - Events

If the deployment worked, you should see that the challenge tried to create one or two "pods" (you can think of a pod as a replica) in Events (at the end). Otherwise, the *Status* will tell you that it wasn't able to do so for some *Reason*. This usually means that the configuration files were manually modified by the author of the task, so it's a good moment to investigate the change history of the files below the `k8s` directory of the challenge directory.

### Looking into the pods

The next step is to look into the replicas of the challenge (the pods). You can list the pods for a specific challenge by running:
```
kubectl get pods --selector=app=kctf-chal-troubleshooting
```

This should tell you the status of the task; a healthy challenge should look like this:
```
NAME                         READY   STATUS    RESTARTS   AGE
apache-php-9877d8b7c-k89zw   2/2     Running   0          24h
apache-php-9877d8b7c-t2p4k   2/2     Running   0          24h
```

Notice that READY says 2/2, which means 2 out of 2 containers are ready. If there's an error, you might see that RESTARTS has a number larger than 0, or READY displays 1/2 or 0/2.

To debug this, run:
```
kubectl describe pods --selector=app=kctf-chal-troubleshooting
```

This will describe the pod; similar to deployment these fields will be of interest:
 - *Status* and *Reason*
 - Events
 - *State* (per container)

A healthy challenge should have "running" as the *Status* and the *State*. Any other Status/State will be explained under *Reason*.

If the *Reason* doesn't make sense, you can find more information under *Events*. Ideally, the last line of *Events* should be:
```
Normal   Created          71m             ...     Created container challenge
```

Otherwise, the Event will explain what the problem was.

## Troubleshooting remotely

Get your hat and your boots ready, because you are now going to be testing on production! This is probably the last resort, and the most likely to not fix any problems (and rather introduce new ones), but this is an unavoidable step, as often errors are easier to debug in production.

### Basic troubleshooting

The commands available for kCTF are similar to the ones available for Docker and kubectl. 

To get the status of a challenge, run:
```
kctf chal status
```

To shell into a challenge, run:
```
kctf chal debug ssh
```

To shell into the healthcheck, run:
```
kctf chal debug ssh --container=healthcheck
```

To obtain remote logs, run:
```
kctf chal debug logs
```

In addition, you can run any kubectl command as described in the section [Inspecting a Kubernetes Deployment](#inspecting-a-kubernetes-deployment).

### Restarting or redeploying

A good first step is to restart the challenge. To do so run:
```
kubectl rollout restart deployment/chal
```
Note: To make Kubernetes automatically restart flaky challenges, you should have a healthcheck. 

To redeploy the challenge (for example, if the challenge works well locally in a local cluster), run:
```
kctf chal start
```
This will deploy the local challenge to the remote cluster. 

To temporarily undo a bad rollout, run:
```
kubectl rollout undo deployment/chal
```
