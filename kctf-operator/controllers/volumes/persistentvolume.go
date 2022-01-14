package volumes

import (
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func persistentVolume(persistentVolumeClaim *corev1.PersistentVolumeClaim,
	challenge *kctfv1.Challenge) *corev1.PersistentVolume {
	// returns persistent volume correspondent to persistentvolumeclaim
	persistentVolume := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      persistentVolumeClaim.Spec.VolumeName,
			Namespace: persistentVolumeClaim.Namespace,
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/disks/gcs/" + challenge.Namespace + "/" +
						challenge.Name + "/" + persistentVolumeClaim.Spec.VolumeName,
				},
			},
			StorageClassName:              "manual",
			Capacity:                      persistentVolumeClaim.Spec.Resources.Requests,
			AccessModes:                   persistentVolumeClaim.Spec.AccessModes,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
		},
	}
	return persistentVolume
}
