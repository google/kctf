package deployment

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Find index of the container with a specific name in a list of containers
func find_idx(name string, containers []corev1.Container) int {
	for i, container := range containers {
		if container.Name == name {
			return i
		}
	}
	return -1
}

// Deployment without Healthcheck
func deployment(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	ls := labelsForChallenge(challenge.Name)
	var replicas int32 = 1
	var readOnlyRootFilesystem = true

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
		},
	}

	if challenge.Spec.PodTemplate != nil {
		deployment.Spec.Template = challenge.Spec.PodTemplate.Template
	}

	// Find the index of container challenge if existent:
	idx_challenge := find_idx("challenge", deployment.Spec.Template.Spec.Containers)

	// if idx_challenge is -1, it means that pod template doesn't contain a container called challenge
	if idx_challenge == -1 {
		// Creates a container called challenge
		challengeContainer := corev1.Container{
			Name: "challenge",
		}
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers,
			challengeContainer)
		idx_challenge = len(deployment.Spec.Template.Spec.Containers) - 1
	}

	// Changes what need to be changed in the Template and in the container challenge
	deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Labels: ls,
		/*Annotations: map[string]string{
			"container.apparmor.security.beta.kubernetes.io/challenge": "localhost/ctf-profile",
		},*/
	}
	// Set container ports based on the ports that were passed
	deployment.Spec.Template.Spec.Containers[idx_challenge].Ports = ContainerPorts(challenge)
	// Set other container's configurations
	deployment.Spec.Template.Spec.Containers[idx_challenge].Image = challenge.Spec.ImageTemplate
	deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext = &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Add: []corev1.Capability{
				"SYS_ADMIN",
			},
		},
		ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
	}

	deployment.Spec.Template.Spec.Containers[idx_challenge].Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(900, resource.DecimalSI),
		},
		Requests: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(450, resource.DecimalSI),
		},
	}

	/*volumeMounts := []corev1.VolumeMount{{
		Name:      "pow",
		ReadOnly:  true,
		MountPath: "/kctf/pow",
	},
		{
			Name:      "pow-bypass-pub",
			ReadOnly:  true,
			MountPath: "/kctf/pow-bypass",
		}}

	deployment.Spec.Template.Spec.Containers[idx_challenge].VolumeMounts =
		append(deployment.Spec.Template.Spec.Containers[idx_challenge].VolumeMounts, volumeMounts...)

	volumes := []corev1.Volume{{
		Name: "pow",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "pow",
				},
			},
		},
	},
		{
			Name: "pow-bypass-pub",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "pow-bypass-pub",
				},
			},
		}}

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volumes...)*/

	return deployment
}
