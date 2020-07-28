// Creates deployment deployment

package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// Deployment with Healthcheck
func (r *ReconcileChallenge) deploymentWithHealthcheck(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	//TODO
	dep := &appsv1.Deployment{}
	return dep
}

// Deployment without Healthcheck
func (r *ReconcileChallenge) deploymentWithoutHealthcheck(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	ls := labelsForChallenge(challenge.Name)
	var replicas int32 = 1
	var readOnlyRootFilesystem = true

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: challenge.Spec.ImageTemplate,
						Name:  "challenge",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{
									"SYS_ADMIN",
								},
							},
							ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
						},
					}}, // TODO: Complete deployment configurations
				},
			},
		},
	}

	// Set container ports based on the ports that were passed
	dep.Spec.Template.Spec.Containers[0].Ports = ContainerPorts(challenge)

	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, r.scheme)
	return dep
}

// deploymentForChallenge returns a challenge Deployment object
func (r *ReconcileChallenge) deploymentForChallenge(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	if challenge.Spec.Healthcheck.Enabled == true {
		return r.deploymentWithHealthcheck(challenge)
	} else {
		return r.deploymentWithoutHealthcheck(challenge)
	}
}

// labelsForChallenge returns the labels for selecting the resources
// belonging to the given challenge CR name.
func labelsForChallenge(name string) map[string]string {
	return map[string]string{"app": "challenge", "challenge_cr": name}
}
