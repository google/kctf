// Creates deployment

package deployment

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ContainerPorts(challenge *kctfv1alpha1.Challenge) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{}

	for _, port := range challenge.Spec.Network.Ports {
		containerPort := corev1.ContainerPort{
			ContainerPort: port.TargetPort.IntVal,
		}
		ports = append(ports, containerPort)
	}

	return ports
}

// labelsForChallenge returns the labels for selecting the resources
// belonging to the given challenge CR name.
func labelsForChallenge(name string) map[string]string {
	return map[string]string{"app": "challenge", "challenge_cr": name}
}

// deploymentForChallenge returns a challenge Deployment object
func Generate(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	if challenge.Spec.Healthcheck.Enabled == true {
		return withHealthcheck(challenge)
	} else {
		return deployment(challenge)
	}
}

func Create(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (reconcile.Result, error) {
	dep := Generate(challenge)
	log.Info("Creating a new Deployment", "Deployment.Namespace",
		dep.Namespace, "Deployment.Name", dep.Name)

	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, scheme)

	err := cl.Create(ctx, dep)

	if err != nil {
		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
			dep.Namespace, "Deployment.Name", dep.Name)
		return reconcile.Result{}, err
	}

	// Deployment created successfully - return and requeue
	return reconcile.Result{Requeue: true}, nil
}
