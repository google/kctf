// File that set values not specified by the user to default
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Functions to return the default values
func PersistentVolumeClaimDefault() corev1.PersistentVolumeClaim {
	var persistentVolumeClaimDefault = corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					// Set 1Gi for the user; TODO: I think this could be done better
					"storage": *resource.NewQuantity(1*1024*1024*1024*1024, resource.BinarySI),
				},
			},
		},
	}
	return persistentVolumeClaimDefault
}

// Function that specifies the Default for PodTemplate
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

	if chal.Spec.Network.Ports == nil {
		chal.Spec.Network.Ports = PortsDefault()
	}

	//TODO: set default values of PersistentVolumeClaim and Deployment
}
