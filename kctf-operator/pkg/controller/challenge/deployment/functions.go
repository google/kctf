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

// Adds volume mounts from podtemplate
// It would be interesting that the user can put some features in the pod template
// and they are kept in the actual deployment, but a StrategicMerge as in kustomization
// wasn't found so we would need to define our own logic for merging
func MergeWithPodTemplate(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) {
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
		append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
			challenge.Spec.PodTemplate.Template.Spec.Containers[0].VolumeMounts...)
}

// labelsForChallenge returns the labels for selecting the resources
// belonging to the given challenge CR name.
func labelsForChallenge(name string) map[string]string {
	return map[string]string{"app": "challenge", "challenge_cr": name}
}

// deploymentForChallenge returns a challenge Deployment object
func DeploymentForChallenge(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	if challenge.Spec.Healthcheck.Enabled == true {
		return deploymentWithHealthcheck(challenge)
	} else {
		return deployment(challenge)
	}
}

func CreateDeployment(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (reconcile.Result, error) {
	dep := DeploymentForChallenge(challenge)
	log.Info("Creating a new Deployment", "Deployment.Namespace",
		dep.Namespace, "Deployment.Name", dep.Name)

	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, scheme)

	err := client.Create(ctx, dep)

	if err != nil {
		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
			dep.Namespace, "Deployment.Name", dep.Name)
		return reconcile.Result{}, err
	}

	// Deployment created successfully - return and requeue
	return reconcile.Result{Requeue: true}, nil
}
