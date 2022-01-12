package deployment

import (
	kctfv1 "github.com/google/kctf/api/v1"
	utils "github.com/google/kctf/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment without Healthcheck
func deployment(challenge *kctfv1.Challenge) *appsv1.Deployment {
	var replicas int32 = 1
	if challenge.Spec.Replicas != nil {
		replicas = *challenge.Spec.Replicas
	}

	var readOnlyRootFilesystem = true

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
			Labels:    map[string]string{"app": challenge.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": challenge.Name},
			},
		},
	}

	if challenge.Spec.PodTemplate != nil {
		deployment.Spec.Template = challenge.Spec.PodTemplate.Template
	}

	// Find the index of container challenge if existent:
	idx_challenge := utils.IndexOfContainer("challenge", deployment.Spec.Template.Spec.Containers)

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
		Labels: map[string]string{"app": challenge.Name},
		Annotations: map[string]string{
			"container.apparmor.security.beta.kubernetes.io/challenge": "unconfined",
		},
	}
	// Set container ports based on the ports that were passed
	deployment.Spec.Template.Spec.Containers[idx_challenge].Ports = containerPorts(challenge)
	// Set other container's configurations
	deployment.Spec.Template.Spec.Containers[idx_challenge].Image = challenge.Spec.Image
	if deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext == nil {
		deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext = &corev1.SecurityContext{}
	}
	if deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.ReadOnlyRootFilesystem == nil {
		deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.ReadOnlyRootFilesystem = &readOnlyRootFilesystem
	}
	if deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.Capabilities == nil {
		deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.Capabilities = &corev1.Capabilities{}
	}

	deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.Capabilities.Add =
		append(deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.Capabilities.Add, "SYS_ADMIN")

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "pow",
			ReadOnly:  true,
			MountPath: "/kctf/pow",
		},
		{
			Name:      "pow-bypass-pub",
			ReadOnly:  true,
			MountPath: "/kctf/pow-bypass",
		},
	}

	deployment.Spec.Template.Spec.Containers[idx_challenge].VolumeMounts =
		append(deployment.Spec.Template.Spec.Containers[idx_challenge].VolumeMounts, volumeMounts...)

	volumes := []corev1.Volume{{
		Name: "pow",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: challenge.Name + "-pow",
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

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volumes...)

	return deployment
}
