// File that set values correctly and return default values that weren't specified
package set

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Function to set the pod template
func setPodTemplate(challenge *kctfv1alpha1.Challenge) {
	// Loop to change things path of the podtemplate
	for i, _ := range challenge.Spec.PodTemplate.Template.Spec.Containers {
		container := &challenge.Spec.PodTemplate.Template.Spec.Containers[i]

		// For each volume mount:
		for j, _ := range container.VolumeMounts {
			volumeMount := &container.VolumeMounts[j]
			// We add where the folder specified should be mounted
			volumeMount.MountPath = "/mnt/disks/" + volumeMount.MountPath
		}
	}
}

// Function to set the persistent volume claim
func setPersistentVolumeClaims(challenge *kctfv1alpha1.Challenge) {
	storageClassName := "manual"

	for i, _ := range challenge.Spec.PersistentVolumeClaims.Items {

		item := &challenge.Spec.PersistentVolumeClaims.Items[i]

		// Setting some configurations
		item.ObjectMeta.Namespace = challenge.Namespace
		item.Spec.StorageClassName = &storageClassName
		item.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			"ReadWriteMany",
		}
		item.Spec.VolumeName = item.ObjectMeta.Name
	}
}

// Function to return the default ports
func portsDefault() []kctfv1alpha1.PortSpec {
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
func DefaultValues(challenge *kctfv1alpha1.Challenge) {
	// Sets default ports
	if challenge.Spec.Network.Ports == nil {
		challenge.Spec.Network.Ports = portsDefault()
	}

	// Sets PodTemplate
	if challenge.Spec.PodTemplate != nil {
		setPodTemplate(challenge)
	}

	// Configure the PersistentVolumeClaim since we don't expect user to pass everything
	if challenge.Spec.PersistentVolumeClaims != nil {
		setPersistentVolumeClaims(challenge)
	}
}
