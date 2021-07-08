# CTF Playbook

Now that your challenges are ready, let's talk about how to set up your cluster to ensure that everything runs smoothly during the CTF event itself.

If you haven't set up a cluster yet, please follow the [google cloud walkthrough](google-cloud.md) to do so.

Also, make sure that every challenge has a working healthcheck. This will allow kubernetes to automatically restart challenges that stop working.

## Scaling

You want to make sure that all the challenges are scaled depending on how much traffic they receive.

First of all, we need to make sure that the cluster has enough nodes (VMs) to run the challenges on.
You can use `kctf cluster resize` to add a new node pool to your cluster and delete the old one.
The parameters allow you to configure the minimum and maximum amount of nodes for automatic scaling as well as what kind of machines to use. For example:

```sh
kctf cluster resize --min-nodes 4 --max-nodes 16 --num-nodes 5 --machine-type n2-standard-4
```

| :warning: the maximum number of nodes may be limited by the [cloud project quotas](https://cloud.google.com/compute/quotas) |
| --- |

After enabling scaling for the number of nodes, we also want to enable scaling for the challenges. You can do this by adding a horizontalAutoScaler spec and a resource request to the `challenge.yaml`.
For available fields, see:

```sh
kubectl explain challenge.spec.horizontalPodAutoscalerSpec
```

An example spec can look like this:

```yaml
apiVersion: kctf.dev/v1
kind: Challenge
metadata:
  name: mychallenge
spec:
  # [...]
  horizontalPodAutoscalerSpec:
    maxReplicas: 8
    minReplicas: 2
    targetCPUUtilizationPercentage: 60
  podTemplate:
    template:
      spec:
        containers:
          - name: 'challenge'
            resources:
              requests:
                memory: "1000Mi"
                cpu: "500m"
```

Start the challenge:

```sh
kctf chal start
```

And you can confirm that the autoscaler was created with `kubectl`:

```sh
kubectl get horizontalPodAutoscaler
```

### Proof of Work

If you notice that one of the challenges uses a high amount of resources, you can enable a proof of work on every connection.
Simply set the `powDifficultySeconds` parameter and restart the challenge:

```yaml
apiVersion: kctf.dev/v1
kind: Challenge
metadata:
  name: mychallenge
spec:
  powDifficultySeconds: 60
```

Note that the proof of work doesn't support web challenges. For those, you can include a captcha on your web endpoints, for example [reCAPTCHA](https://www.google.com/recaptcha/about/).

## Monitoring

We all know IRC is the best way to get alerts about broken challenges :). But if you want to keep an eye on the challenges and catch potential issues early, there are a few ways to check the health.

First of all, you can list all challenges with `kubectl`, this will show you which are deemed healthy by the healthcheck and which are not. Remember, healthchecks are important!

```sh
$ kubectl get challenges
NAME         HEALTH      STATUS    DEPLOYED   PUBLIC
mychallenge  healthy     Running   true       true
$ cd mychallenge
$ kctf chal status
```

If any of the challenges are broken, you can check out our [troubleshooting docs](troubleshooting.md) for some debugging tips.

Another option is to use the [google cloud web UI](https://console.cloud.google.com), which shows you various information about your cluster. For example:
* [Clusters](https://console.cloud.google.com/kubernetes/list) includes how many nodes are currently running.
* [Workloads](https://console.cloud.google.com/kubernetes/workload) has data on CPU/Memory/Disk usage of every challenge.
* [Monitoring](https://console.cloud.google.com/monitoring) allows you to create dashboards and set up alerts.
