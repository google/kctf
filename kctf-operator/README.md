# kCTF Operator

## Note for developers

The operator is called automatically by the scripts but, in case you want to try something on your own:

This operator was created using Operator-SDK 0.18.

Tutorial on how to use Operator-SDK 0.18: https://v0-18-x.sdk.operatorframework.io/docs/golang/quickstart/

The Custom Resource Definition is created inside the folder "deploy/crds" and it's generated from the file "pkg/apis/kctf/v1alpha1/challenge_types.go".

The operator logic is all inside "pkg/controller/challenge/", where the main file is "challenge_controller.go".

Before running the operator, you have to register the CRD in the server by creating it: "kubectl create -f deploy/crds/kctf.dev\_challenges\_crd.yaml".

There are two ways of making the operator work:

1. You can run locally by running "operator-sdk run local --watch-namespace="" ". 

2. You can run in the cluster by running "kubectl create -f deploy/rbac.yaml" and "kubectl create -f deploy/operator.yaml".

To create a challenge, you must run "kubectl apply -f /path/to/cr\_of\_the\_challenge.yaml". If you change it, you should run the same command to apply the changes.

