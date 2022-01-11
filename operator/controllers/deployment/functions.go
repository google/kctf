// Creates deployment

package deployment

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isEqual(deploymentFound *appsv1.Deployment,
	deployment *appsv1.Deployment) bool {
	return reflect.DeepEqual(deploymentFound.Spec.Template.Spec,
		deployment.Spec.Template.Spec)
}

func containerPorts(challenge *kctfv1.Challenge) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{}

	for _, port := range challenge.Spec.Network.Ports {
		containerPort := corev1.ContainerPort{
			ContainerPort: port.TargetPort.IntVal,
		}
		ports = append(ports, containerPort)
	}

	return ports
}

// deploymentForChallenge returns a challenge Deployment object
func generate(challenge *kctfv1.Challenge) *appsv1.Deployment {
	if challenge.Spec.Healthcheck.Enabled == true {
		return withHealthcheck(challenge)
	} else {
		return deployment(challenge)
	}
}

func create(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	dep := generate(challenge)
	log.Info("Creating a new Deployment", "Deployment.Namespace",
		dep.Namespace, "Deployment.Name", dep.Name)

	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, scheme)

	err := cl.Create(ctx, dep)

	if err != nil {
		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
			dep.Namespace, "Deployment.Name", dep.Name)
		return false, err
	}

	// Deployment created successfully - return and requeue
	return true, nil
}

func Update(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Flags if there was a change
	change := false

	deploymentFound := &appsv1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, deploymentFound)

	// Just enters here if it's a new deployment
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		return create(challenge, client, scheme, log, ctx)

	} else if err != nil {
		log.Error(err, "Couldn't get the deployment", "Challenge Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}

	// Checks if the deployment is correctly set
	if dep := generate(challenge); !isEqual(deploymentFound, dep) {
		change = true
		deploymentFound.Spec.Template.Spec = dep.Spec.Template.Spec
	}

	// Ensure if the challenge is ready and, if not, set replicas to 0
	changedReplicas := updateNumReplicas(deploymentFound.Spec.Replicas, challenge)
	// Ensure that the images of the challenge and of the healthcheck are the same as the ones in the CR
	changedImage := updateImages(deploymentFound, challenge)

	// Checks if there was a change in the deployment
	change = change || changedReplicas || changedImage

	// Updates deployment with client
	if change == true {
		err = client.Update(ctx, deploymentFound)
		if err != nil {
			log.Error(err, "Failed to update deployment", "Challenge Name: ",
				challenge.Name, " with namespace ", challenge.Namespace)
			return false, err
		}
		log.Info("Deployment updated succesfully", "Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return true, nil
	}

	return false, nil
}
