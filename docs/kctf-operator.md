this document is outdated

Please check

https://github.com/google/kctf/issues/356

---

# Developing the operator

## Introduction

The kctf-operator is responsible for deploying the Kubernetes configurations in the cluster and 
keeps everything up-to-date accordingly to the configuration we put in the Custom Resource specific to the challenge.

This operator is created automatically in the scripts when you run `start.sh`. 
In case you want to run it alone, go directly to [Testing locally](#https://github.com/google/kctf/blob/beta/docs/kctf-operator.md#testing-locally) or [Deploying the operator](https://github.com/google/kctf/blob/beta/docs/kctf-operator.md#deploying-the-operator). They correspond, respectively, to local testing and cluster testing.

This code was implemented using operator-sdk 0.18, so ensure you have it installed. 
If you want to know more about what you can do using operator-sdk, you can access: https://v0-18-x.sdk.operatorframework.io/docs/golang/quickstart/.

## Changing the code

About the structure of the code, inside the folder deploy, we have `operator.yaml`, 
`rbac.yaml` and a folder called `crds`. The first one is the yaml file of the operator, 
which creates its deployment. The second one is the necessary permissions that the operator need to run. 
Finally, the third one is where the CRDs are stored when you generate them.

We have also the folder `pkg`, which contains three other folders: `apis`, `controller` and `resources`. 
The `apis `folder contains the code responsible for generating the CRDs and for Deep Copy. 
The important file there is `challenge_types.go`, which is inside `pkg/apis/kctf/v1alpha1/`. 
It defines the specifications and the status of the challenge. Using kubebuilder in the comments, 
you can also define other things such as if the field is mandatory or not.
Inside the `controller` folder, we have the code of the operator logic. Inside `pkg/controller/challenge`, 
you have the packages for each utility and the file `challenge_controller.go`
contains the Reconcile function, which is called every time there's a change in the watched objects. 
The `resources` folder contains objects that
are created when the operator is initialized. These resources consist of the services that kCTF provides.

The folder `cmd/manager` contains the `main.go`, which is the main function of the code and the folder `build` contains necessary code so that 
everything works right. This last folder shouldn't be modified.

Finally, the folder `samples` contains some CRs as example and the folder `version` contains the current version of the operator.

## Generating CRD and the Deep Copy code

In order to generate a new CRD based on the new code in `challenge_types.go`, you have to run from the `kctf-operator` folder:

```
operator-sdk generate crds
```

After generating the CRD, have a look in the issue #136. Don't forget to apply/create the CRD after generating it.

And, to update `zz_generated.deepcopy.go`:

```
operator-sdk generate k8s
```

## Testing locally

You can run the operator locally and see the logs that come from it by running:

```
operator-sdk run local --watch-namespace=""
```

## Creating an image, pushing it and changing the `operator.yaml`

You can create a new image considering your changes in the code by running:

```
operator-sdk build gcr.io/myrepo/myimagename:tag
```

And, you can push it to your repository by doing:

```
docker push gcr.io/myrepo/myimagename:tag
```

Remember to change the image in `deploy/operator.yaml` to make the operator use this image.

## Deploying the operator

You can deploy the operator by running:

```
kubectl apply -f deploy/operator.yaml
```

## Testing with sample custom resources

You can test the operator by applying the samples CRs:

```
kubectl apply -f samples/mychal.yaml
```

## More information

You can find more information in the website of operator-sdk cited above, where it says how to generate new CRDs of other versions and how to create controllers specific to them.
