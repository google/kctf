# Advanced Task Configuration

> **Watch out!** changing things here is meant to be done only for some advanced use-cases.\
> Don't touch things here unless you know what you are doing.

This directory allows you to configure each task individually.
If you wish to make changes that apply to all tasks, edit the files in `kctf-conf/base` of your challenge directory instead.

## `network/network.yaml`

This file is used to configure the [load balancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) that exposes a task to the internet.
You only need to touch this file if you want to expose your task in a different port (like port 80).

> **Note**: If your challenge needs more than one port (in addition to port `1337`), you must also edit [`containers.yaml`](#deploymentcontainersyaml).

These are the contents of the `network/network.yaml` file, without any changes:
```yaml
apiVersion: "v1"
kind: "Service"
metadata:
  name: "chal"
```

If you want, for example, to expose port `1337` on port `80`, then you would configure that in the same way as done for the `apache-php` sample task [`/samples/apache-php/config/advanced/network/network.yaml`](https://github.com/google/kctf/blob/master/samples/apache-php/config/advanced/network/network.yaml):
```yaml
apiVersion: "v1"
kind: "Service"
metadata:
  name: "chal"
spec:
  type: "LoadBalancer"
  ports:
  - name: "http"
    protocol: "TCP"
    port: 80
    targetPort: 1337
```

> **Note**: By default, port `1337` of the nsjail container is exposed in port `1` of the public IP of the task,
you don't have to configure that, it's already done for all tasks in kCTF automatically.

## `deployment/autoscaling.yaml`

This file is used to configure the [autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/), which allows you to configure how many replicas of a task are deployed.
You only need to touch this file if you want to limit the number of replicas (minimum or maximum), or to configure the way or metrics used to scale it.

These are the contents of the `deployment/autoscaling.yaml` file, without any changes:
```yaml
apiVersion: "autoscaling/v1"
kind: "HorizontalPodAutoscaler"
metadata:
  name: "chal"
```

If you want, for example, to set a minimum and maximum number of tasks to be `2`, then you would configure that in the same way as done for the `apache-php` sample task [`/samples/apache-php/config/advanced/deployment/autoscaling.yaml`](https://github.com/google/kctf/blob/master/samples/apache-php/config/advanced/deployment/autoscaling.yaml):
```yaml
apiVersion: "autoscaling/v1"
kind: "HorizontalPodAutoscaler"
metadata:
  name: "chal"
spec:
  minReplicas: 2
  maxReplicas: 2
```

> **Note**: By default, the maximum number of replicas is 1, which means **autoscaling is disabled by default**.

## `deployment/containers.yaml`

This file is used to configure the [deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), which includes among other things the configuration of the task containers.
You only need to touch this file if you want to add another port (in addition to `1337`) for your task, or if you want to have a shared directory between task replicas.

These are the contents of the `deployment/containers.yaml` file, without any changes:
```yaml
apiVersion: "apps/v1"
kind: "Deployment"
metadata:
  name: "chal"
```

If you want, for example, to share a folder (backed by Google Cloud Storage) among all replicas of a task, then you would configure that in the same way as done for the `apache-php` sample task [`/samples/apache-php/config/advanced/deployment/containers.yaml`](https://github.com/google/kctf/blob/master/samples/apache-php/config/advanced/deployment/containers.yaml):
```yaml
apiVersion: "apps/v1"
kind: "Deployment"
metadata:
  name: "chal"
spec:
  template:
    spec:
      containers:
      - name: challenge
        volumeMounts:
        - name: sessions
          mountPath: /mnt/disks/sessions
        - name: uploads
          mountPath: /mnt/disks/uploads
      volumes:
      - name: sessions
        hostPath:
          path: /mnt/disks/gcs/apache-php/sessions
      - name: uploads
        hostPath:
          path: /mnt/disks/gcs/apache-php/uploads
```

> **Note**: The `/mnt/disks/gcs/` directory is mapped to a GCS bucket shared by the whole cluster, so you **MUST** only expose a subdirectory in order to keep different tasks isolated.

To expose another port (in addition to `1337`), specify:
```yaml
apiVersion: "apps/v1"
kind: "Deployment"
metadata:
  name: "chal"
spec:
  template:
    spec:
      containers:
      - name: challenge
        ports:
        - containerPort: 1234
```

> **Note**: For the additional port to be visible on the public IP of the cluster, you must also edit the load balancer in [`network.yaml`](#networknetworkyaml).
