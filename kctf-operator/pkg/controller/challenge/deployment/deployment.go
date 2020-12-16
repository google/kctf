package deployment

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	utils "github.com/google/kctf/pkg/controller/challenge/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment without Healthcheck
func deployment(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	var replicas int32 = 1
	if challenge.Spec.Replicas != nil {
		replicas = *challenge.Spec.Replicas
	}

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
			"container.apparmor.security.beta.kubernetes.io/challenge": "localhost/ctf-profile",
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
		var readOnlyRootFilesystem = true
		deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.ReadOnlyRootFilesystem = &readOnlyRootFilesystem
	}
	if deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.ProcMount == nil {
		procMountType := corev1.UnmaskedProcMount
		deployment.Spec.Template.Spec.Containers[idx_challenge].SecurityContext.ProcMount = &procMountType
	}

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}
	if deployment.Spec.Template.Spec.SecurityContext.RunAsUser == nil {
		var uid int64 = 1000
		deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &uid
	}

	deployment.Spec.Template.Spec.Containers[idx_challenge].Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(900, resource.DecimalSI),
		},
		Requests: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(450, resource.DecimalSI),
		},
	}

	volumeMounts := []corev1.VolumeMount{{
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

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volumes...)

	return deployment
}
