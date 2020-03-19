# kCTF Infrastructure Walkthrough

Welcome to the kCTF walkthrough for Google Cloud!

## Goal of this walkthrough

The purpose of this walkthrough is to guide you through the configuration of the kCTF infrastructure using Google Cloud.

Note: If not already doing so, you can also open this walkthrough directly in [Google Cloud Shell](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/google/kctf&tutorial=docs/google-cloud.md).

# Step 1 – Set up the cluster

The first step consists of setting up the cluster and the related infrastructure.

## Enable the necessary GCP features
Set up billing, and enable the compute API:
1. <walkthrough-project-billing-setup>Select a project and enable billing.</walkthrough-project-billing-setup>
1. <walkthrough-enable-apis apis="compute.googleapis.com,container.googleapis.com,containerregistry.googleapis.com">Enable the compute API.</walkthrough-enable-apis>

You can enable APIs from the command line with:
```
gcloud services enable compute container containerregistry.googleapis.com
gcloud services enable compute container compute.googleapis.com
gcloud services enable compute container container.googleapis.com
```

## Configure the project

Perform the following steps to configure your project.

### Make sure your umask allows copying executable files

```
umask 0022
```

### Enable docker integration with Google Container Registry

```
gcloud auth configure-docker
```

### Add the bin directory to your PATH

```
PATH="$PATH:$(pwd)/bin"
```

### Run the configuration script

```
kctf-config-create --chal-dir ~/kctf-demo --project {{project-id}} --start-cluster
```

### Set configuration properties
Enter a path for storing the challenges:
```
~/kctf-demo
```

Enter your project id:
```
{{ project-id }}
```

### Other settings 
For all other settings, use the default values.

## Create the cluster
After configuring the project, the cluster is created automatically. This is only done once per CTF. A "cluster" is essentially a group of VMs with a "master" that defines what to run there.

It can take around a minute for the cluster to be created, during which the following message is displayed:

```Creating cluster ... in ... Cluster is being health-checked...``` 

You'll get a message when cluster creation has completed.

While you wait, here's how the infrastructure works:
1. The CTF challenges run inside "nsjail" (a security sandbox).
1. The contents of the nsjail sandbox are configured in a docker container.
1. The container is deployed using Kubernetes in a group of VMs.
1. If the VMs consume too much CPU, Kubernetes automatically deploys more VMs.
1. If the VMs consume too little CPU, Kubernetes will shut down some VMs.
1. Some very expensive challenges can be allocated to the same VMs as less busy challenges.

The above steps ensure the availability of the challenges, while using computing resources in a resourceful manner.

Your cluster should soon be ready, when it is, you can continue with this walkthrough.

Note: If you are curious and have some spare time, take a look at the [kCTF introduction](https://github.com/google/kctf/blob/master/docs/introduction.md), which includes a quick 8 minute summary of what kCTF is and how it interacts with Kubernetes.

# Step 2 – Create a challenge
Now that you have set up a cluster, you can create a challenge.

In this step, you'll learn how to:
* Create a challenge called "demo-challenge"
* Build a Docker image
* Deploy the Docker image to the cluster
* Expose the Docker image to the internet

Note: The cluster must be created before you continue, otherwise the following commands won't work. To create a cluster, see the preceding steps. Continue with the next steps if you already created a cluster.

## Call kctf-chal-create.sh to copy the skeleton
Run the following command to create a challenge called "demo-challenge":
```
kctf-chal-create demo-challenge
```

This creates a directory called `demo-challenge` under the `kctf-demo` directory.

If you look inside `demo-challenge`, you can see the challenge configuration. The file in `challenge/image/chal` is the entry-point, which means that it is executed every time a TCP connection is established. While this demo challenge just prints the flag, a real challenge would instead expose a an actual challenge.

In the next step you'll find out how to create a docker image with the newly created challenge.

## Deploy the challenge

To deploy the challenge, run the following command, which builds and deploys the challenge, **but doesn't expose it to the internet**:

```
cd ~/kctf-demo/demo-challenge
make start
```

This command deploys the image to your cluster, which will soon start consuming CPU resources. The challenge automatically scales based on CPU usage.

## Expose your challenge to the internet
Run the following command to create a new file:

```
emacs ~/kctf-demo/demo-challenge/config/chal.conf
```

Modify the file by entering `PUBLIC=true`, then run the following command:

```
make start
```

This step can take a minute. It reserves an IP address for your challenge and redirects traffic to your docker containers when someone connects to it. Wait for it to finish before continuing.

While you wait, some important information to be aware of:
 * You should only expose your challenge to the internet once the challenge is ready to be released to the public. Don't expose your challenge too early or the challenge will leak.
 * The ports exposed by the challenge are configured by nsjail (see `nsjail.cfg`) and `config/advanced/network/network.yaml`. Make sure these files are kept in sync.
 * In `network.yaml`, `targetPort` is the port that nsjail is listening on. `port` is the port that the external IP listens on.

# Step 3 – Test the challenge

Now that you have a challenge up and running, you need to test it to make sure it works. In this step, you will:
* Connect to the challenge
* Add and configure a proof of work, and update the running task
* Learn how to debug Kubernetes

## Connect to the challenge

Run the following command to connect to your challenge:

```
nc $(make ip) 1
```

If all went well, you should see a shell. 

Debugging failures here is easy, here are some things you could do if this didn't work:
1. Go to [Services in GKE](https://console.cloud.google.com/kubernetes/discovery)
1. Select demo-challenge
1. Under *Stackdriver Logs* click on demo-challenge

If there were any errors deploying the challenge, they should be visible here.

In the next step we'll see how to edit the challenge, add a proof of work to prevent abuse, and push an update.

## Add a proof of work
To add a proof of work, edit the configuration of the challenge in `config/pow.conf`:

1. Open <walkthrough-editor-select-regex filePath="kctf-demo/demo-challenge/config/pow.conf" regex="0">pow.yaml</walkthrough-editor-select-regex> and change the 0 to 1.
1. Run the following command to enable the proof of work:
    ```
    make start
    ```

Note: This is a very weak proof of work (strength of 1). For it to be useful in a real CTF, you probably want to set it to 15, 20, or more to require people to actually do some work. That said, for this walkthrough, let's take it easy, and leave it at 1.

Once the challenge is updated, run:
```
nc $(kctf-chal-ip demo-challenge) 1
```

This connects you to the challenge with a proof of work in front. Enter **00** to pass the proof of work (as mentioned, the difficulty of 1 isn't very strict). If it doesn't work, try again (or run the script).

And that's it. Now that you saw how to push a challenge and update it, let's see how you can debug the Kubernetes deployment.

## Inspect the Kubernetes deployment
To debug the Kubernetes deployment, you can use `kctf-kubectl` with [kubectl commands](https://kubernetes.io/docs/reference/kubectl/cheatsheet/).

# Step 4 – Clean the challenge
This is the last step of this walkthrough. Performing this step helps save resources after the end of the CTF.

## Delete the challenge
To delete a challenge, run:
```
make stop
```

To test the deletion, run:
```
telnet $(kctf-chal-ip demo-challenge) 1
```

If deletion worked correctly, the connection will fail and an error is returned instead.

## Stop the cluster
To avoid being charged for the VMs any longer, stop the cluster by running:
```
kctf-cluster-stop
```

Note: This stops the cluster, and all the challenges with it. Only stop the cluster if you really want to stop it permanently.

Thank you for completing this walkthrough, and good luck with your CTF challenges!
