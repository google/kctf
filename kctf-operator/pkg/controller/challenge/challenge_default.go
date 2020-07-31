// File that set values not specified by the user to default
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Functions to return the default values
func PersistentVolumeClaimsDefault(challenge *kctfv1alpha1.Challenge) *corev1.PersistentVolumeClaimList {
	stor, _ := resource.ParseQuantity("1Gi")
	var persistentVolumeClaimsDefault = &corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: challenge.Name + "-sessions",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": stor,
						},
					},
				},
			},
		},
	}
	return persistentVolumeClaimsDefault
}

// Function to return the default for PodTemplate
func PodTemplateDefault(challenge *kctfv1alpha1.Challenge) *corev1.PodTemplate {
	var podTemplateDefault = &corev1.PodTemplate{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "challenge",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "sessions",
						MountPath: "/mnt/disks/sessions",
					}},
				}},
				Volumes: []corev1.Volume{{
					Name: "sessions",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: challenge.Name + "-sessions",
						},
					},
				}},
			},
		},
	}
	return podTemplateDefault
}

func SetPersistentVolumeClaims(challenge *kctfv1alpha1.Challenge) {
	storageClassName := "manual"

	for i, _ := range challenge.Spec.PersistentVolumeClaims.Items {

		Item := &challenge.Spec.PersistentVolumeClaims.Items[i]

		// Setting some configurations
		Item.ObjectMeta.Namespace = challenge.Namespace
		Item.Spec.StorageClassName = &storageClassName
		Item.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			"ReadWriteMany",
		}
		Item.Spec.VolumeName = Item.ObjectMeta.Name
	}
}

// Function to return the default ports
func PortsDefault() []kctfv1alpha1.PortSpec {
	var portsDefault = []kctfv1alpha1.PortSpec{
		kctfv1alpha1.PortSpec{
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
func SetDefaultValues(challenge *kctfv1alpha1.Challenge) {
	// Set default ports
	if challenge.Spec.Network.Ports == nil {
		challenge.Spec.Network.Ports = PortsDefault()
	}

	// Set default PodTemplate
	// To verify if the PodTemplate is empty, we check if there aren't any containers
	if challenge.Spec.PodTemplate == nil {
		challenge.Spec.PodTemplate = PodTemplateDefault(challenge)
	}

	// Set default PersistentVolumeClaim
	// To verify if the PersistentVolumeClaim wasn't defined, we check if the volume name is empty
	if challenge.Spec.PersistentVolumeClaims == nil {
		challenge.Spec.PersistentVolumeClaims = PersistentVolumeClaimsDefault(challenge)
	}
	// Configure the PersistentVolumeClaim since we don't expect user to pass everything
	SetPersistentVolumeClaims(challenge)
}
