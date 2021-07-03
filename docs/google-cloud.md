# kCTF Infrastructure Walkthrough

<!-- {% assign project-id = "your_project_id" %} -->

Welcome to the kCTF walkthrough for Google Cloud!

## Goal of this walkthrough

The purpose of this walkthrough is to guide you through the configuration of the kCTF infrastructure using Google Cloud.

  **Note**: If not already doing so, you can also open this walkthrough directly in [Google Cloud Shell](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/google/kctf&tutorial=docs/google-cloud.md&shellonly=true).

# Step 1 – Set up the environment

The first step consists of setting up the cluster and the related infrastructure.

## Enable the necessary GCP features
Set up billing, and enable the compute API:
1. <walkthrough-project-billing-setup>Select a project and enable billing.</walkthrough-project-billing-setup>
1. <walkthrough-enable-apis apis="compute.googleapis.com,container.googleapis.com,containerregistry.googleapis.com,dns">Enable the compute API.</walkthrough-enable-apis>

## Configure the project

Perform the following steps to configure your project.

### Make sure your umask allows copying executable files

```bash
umask 0022
```

### Enable docker integration with Google Container Registry

```bash
gcloud auth configure-docker
```

### Install netcat

```bash
sudo apt install netcat
```

## Setup kCTF

### Download and activate kCTF
```bash
mkdir kctf-demo && cd kctf-demo
curl -sSL https://kctf.dev/sdk | tar xz
source kctf/activate
```

If you have not done this already, you should enable APIs with:
```bash
gcloud services enable compute container containerregistry.googleapis.com dns
```

### Create the GKE cluster
```bash
kctf cluster create --project {{project-id}} --domain-name {{project-id}}-codelab.kctf.cloud --start remote-cluster
```

Note that this will register a domain name for you (`{{project-id}}-codelab.kctf.cloud`) – this means your project ID will be public, and all challenge names you add will be public as well (through DNS records). Every challenge you receive gets mapped to the challenge backend with DNS.

After configuring the project, the cluster is created automatically. This is only done once per CTF. A "cluster" is essentially a group of VMs with a "master" that defines what to run there.

It can take around 5 minutes for the cluster to be created, during which the following message is displayed:

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

  **Note**: If you are curious and have some spare time, take a look at the [kCTF introduction](https://google.github.io/kctf/introduction.html), which includes a quick 8 minute summary of what kCTF is and how it interacts with Kubernetes.

# Step 2 – Create a challenge
Now that you have set up a cluster, you can create a challenge.

In this step, you'll learn how to:
* Create a challenge called "demo-challenge"
* Build a Docker image
* Deploy the Docker image to the cluster
* Expose the Docker image to the internet

**Note**: The cluster must be created before you continue, otherwise the following commands won't work. To create a cluster, see the preceding step. Continue with the next steps if you already created a cluster.

## Call kctf chal create to copy the skeleton
Run the following command to create a challenge called "demo-challenge":
```bash
kctf chal create demo-challenge && cd demo-challenge
```

This will create a default `pwn`-style challenge. We also have `web` and `xss-bot` template challenges which you can select with the `--template` parameter.

This creates a directory called `demo-challenge` under the `kctf-demo` directory.

If you look inside `demo-challenge`, you can see the challenge configuration. While this demo challenge just prints the flag, a real challenge would instead expose a more complex service.

In the next step, you'll find out how to deploy the newly created challenge.

## Deploy the challenge

To deploy the challenge, run the following command, which builds and deploys the challenge, **but doesn't expose it to the internet**:

```bash
make -C challenge && kctf chal start
```

This command deploys the image to your cluster, which will soon start consuming CPU resources. The challenge automatically scales based on CPU usage.

Note that the pwn template comes with a Makefile to build the challenge binary.
This is recommended if you want to hand out the binary as an attachment to
players, e.g. since the layout might matter for ROP gadgets. If the layout
doesn't matter, you could also build it in an intermediate container as part
of your Dockerfile.

## Expose your challenge to the internet

In order to expose your challenge to the internet, you must mark it as public. To do so, edit the `challenge.yaml` file, or run the command below:
```bash
sed -i s/public:\ false/public:\ true/ challenge.yaml
```

Then run the following command:

```bash
kctf chal start
```

This step can take a minute. It reserves an IP address for your challenge and redirects traffic to your docker containers when someone connects to it. Wait for it to finish before continuing.

While you wait, some important information to be aware of:
 * You should only expose your challenge to the internet once the challenge is ready to be released to the public. Don't expose your challenge too early or the challenge will leak.
 * The port exposed by the challenge is configured by nsjail (see `challenge/nsjail.cfg`). **By default, the internal nsjail port 1337 is exposed externally**.  For testing, you can use the `kctf chal debug port-forward` command to connect to it.

# Step 3 – Test the challenge

Now that you have a challenge up and running, you need to test it to make sure it works. In this step, you will:
* Connect to the challenge
* Add and configure a proof of work, and update the running task
* Learn how to debug Kubernetes

## Connect to the challenge

Run the following command to connect to your challenge:

```bash
nc demo-challenge.{{project-id}}-codelab.kctf.cloud 1337
```

If all went well, you should see the flag.

Debugging failures here is easy, here are some things you could do if this didn't work:
1. Go to [Services in GKE](https://console.cloud.google.com/kubernetes/discovery)
1. Select demo-challenge
1. Under *Stackdriver Logs* click on demo-challenge

If there were any errors deploying the challenge, they should be visible here.

In the next step we'll see how to edit the challenge, add a proof of work to prevent abuse, and push an update.

## Add a proof of work
To add a proof of work, edit the configuration of the challenge:

1. Open `challenge.yaml` and change `powDifficultySeconds` from 0 to 1.
    ```bash
    emacs challenge.yaml
    ```
1. Run the following command to enable the proof of work:
    ```bash
    kctf chal start
    ```

  **Note**: This is a very weak proof of work (strength of 1 second). For it to be useful in a real CTF, you probably want to set it to 10 seconds of work, or more. That said, for this walkthrough, let's take it easy, and leave it at 1.

Once the challenge is updated, run:
```bash
nc demo-challenge.{{project-id}}-codelab.kctf.cloud 1337
```

This connects you to the challenge with a proof of work in front.

And that's it. Now that you saw how to push a challenge and update it, let's see how you can debug the Kubernetes deployment.

# Step 4 – Clean the challenge
This is the last step of this walkthrough. Performing this step helps save resources after the end of the CTF.

## Delete the challenge
To delete a challenge, run:
```bash
kctf chal stop
```

### Stop the cluster
To avoid being charged for the VMs any longer, stop the cluster by running:
```bash
kctf cluster stop
```

  **Note**: This stops the cluster, and all the challenges with it. Only stop the cluster if you really want to stop it permanently.

Thank you for completing this walkthrough, and good luck with your CTF challenges!
