// Creates persistentVolumeClaims
package volumes

import (
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func persistentVolumeClaim(claim string,
	challenge *kctfv1.Challenge) *corev1.PersistentVolumeClaim {
	storageClassName := "manual"
	requirement, _ := resource.ParseQuantity("10Gi")
	resources := map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: requirement}

	// returns persistent volume correspondent to persistentvolumeclaim
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claim,
			Namespace: challenge.Namespace,
			Labels: map[string]string{
				"app": challenge.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteMany",
			},
			VolumeName: claim,
			Resources: corev1.VolumeResourceRequirements{
				Requests: resources,
			},
		},
	}
	return persistentVolumeClaim
}
