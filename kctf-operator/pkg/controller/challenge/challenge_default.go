// File that set values not specified by the user to default
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Function to return the default values of autoscaling
func AutoscalingDefault(chal *kctfv1alpha1.Challenge) kctfv1alpha1.AutoscalingSpec {
	var AutoscalingDefault = kctfv1alpha1.AutoscalingSpec{
		MinReplicas:                    1,
		MaxReplicas:                    1,
		TargetCPUUtilizationPercentage: 50,
	}
	return AutoscalingDefault
}

// Functions to return the default values
func PersistentVolumeClaimsDefault(chal *kctfv1alpha1.Challenge) corev1.PersistentVolumeClaimList {
	var persistentVolumeClaimsDefault = corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: chal.Name,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							// Set 1Gi for the user
							"storage": *resource.NewQuantity(1*1024*1024*1024*1024, resource.BinarySI),
						},
					},
				},
			},
		},
	}
	return persistentVolumeClaimsDefault
}

// Function to return the default for PodTemplate
// TODO: implement this
func PodTemplateDefault() corev1.PodTemplate {
	var podTemplateDefault = corev1.PodTemplate{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				InitContainers: nil,
				Containers:     nil,
			},
		},
	}
	return podTemplateDefault
}

// Function to return the default ports
func PortsDefault() []corev1.ServicePort {
	var portsDefault = []corev1.ServicePort{
		corev1.ServicePort{
			// Keeping the same name as in previous network file
			Name:       "netcat",
			Port:       1,
			TargetPort: intstr.FromInt(1337),
			Protocol:   "TCP",
		},
	}
	return portsDefault
}

// Function to check if all is set to default
func SetDefaultValues(chal *kctfv1alpha1.Challenge) {
	// Set default ports
	if chal.Spec.Network.Ports == nil {
		chal.Spec.Network.Ports = PortsDefault()
	}

	// Set default PodTemplate
	// To verify if the PodTemplate is empty, we check if there aren't any containers
	if chal.Spec.PodTemplate.Template.Spec.Containers == nil {
		chal.Spec.PodTemplate = PodTemplateDefault()
	}

	// Set default PersistentVolumeClaim
	// To verify if the PersistentVolumeClaim wasn't defined, we check if the volume name is empty
	if chal.Spec.PersistentVolumeClaims.Items == nil {
		chal.Spec.PersistentVolumeClaims = PersistentVolumeClaimsDefault(chal)
	}

	// Set default of autoscaling
	// To verify if the autoscaling wasn't defined, we check MaxReplicas
	if chal.Spec.Autoscaling.MaxReplicas == 0 {
		chal.Spec.Autoscaling = AutoscalingDefault(chal)
	}
}
