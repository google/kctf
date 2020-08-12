package deployment

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment with Healthcheck
func deploymentWithHealthcheck(challenge *kctfv1alpha1.Challenge,
	scheme *runtime.Scheme) *appsv1.Deployment {
	dep := deployment(challenge, scheme)

	// Get the container with the challenge and add healthcheck configurations
	dep.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		FailureThreshold: 2,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(45281),
			},
		},
		InitialDelaySeconds: 45,
		TimeoutSeconds:      3,
		PeriodSeconds:       30,
	}

	dep.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(45281),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
	}

	healthcheckContainer := corev1.Container{
		Name:    "healthcheck",
		Image:   "healthcheck",
		Command: []string{},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"cpu": *resource.NewMilliQuantity(1000, resource.DecimalSI),
			},
			Requests: corev1.ResourceList{
				"cpu": *resource.NewMilliQuantity(50, resource.DecimalSI),
			},
		},
		// Uncomment when start testing with real challenges
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "pow-bypass",
			ReadOnly:  true,
			MountPath: "/pow-bypass",
		}},
	}

	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, healthcheckContainer)

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
