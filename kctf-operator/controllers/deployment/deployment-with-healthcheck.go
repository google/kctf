package deployment

import (
	kctfv1 "github.com/google/kctf/api/v1"
	utils "github.com/google/kctf/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment with Healthcheck
func withHealthcheck(challenge *kctfv1.Challenge) *appsv1.Deployment {
	dep := deployment(challenge)

	idx_challenge := utils.IndexOfContainer("challenge", dep.Spec.Template.Spec.Containers)
	idx_healthcheck := utils.IndexOfContainer("healthcheck", dep.Spec.Template.Spec.Containers)

	challengeContainer := &dep.Spec.Template.Spec.Containers[idx_challenge]

	// Get the container with the challenge and add healthcheck configurations
	livenessProbe := &challengeContainer.LivenessProbe
	if *livenessProbe == nil {
		*livenessProbe = &corev1.Probe{
			FailureThreshold: 2,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(45281),
				},
			},
			InitialDelaySeconds: 45,
			TimeoutSeconds:      3,
			PeriodSeconds:       30,
		}
	}

	readinessProbe := &challengeContainer.ReadinessProbe
	if *readinessProbe == nil {
		*readinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(45281),
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
		}
	}

	if idx_healthcheck == -1 {
		healthcheckContainer := corev1.Container{
			Name: "healthcheck",
		}
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, healthcheckContainer)
		idx_healthcheck = len(dep.Spec.Template.Spec.Containers) - 1
	}

	healthcheckContainer := &dep.Spec.Template.Spec.Containers[idx_healthcheck]

	healthcheckContainer.Image = challenge.Spec.Healthcheck.Image
	healthcheckContainer.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(1000, resource.DecimalSI),
		},
		Requests: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(50, resource.DecimalSI),
		},
	}

	healthcheckContainer.VolumeMounts = []corev1.VolumeMount{{
		Name:      "pow-bypass",
		ReadOnly:  true,
		MountPath: "/pow-bypass",
	}}

	healthcheckVolume := corev1.Volume{
		Name: "pow-bypass",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "pow-bypass",
			},
		},
	}

	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, healthcheckVolume)

	return dep
}
