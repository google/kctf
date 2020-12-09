package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewServiceAccountGcsFuseSa() runtime.Object {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcsfuse-sa",
			Namespace: "kctf-system",
		},
	}
	return serviceAccount
}

func NewServiceAccountExternalDnsSa() runtime.Object {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-dns-sa",
			Namespace: "kctf-system",
		},
	}
	return serviceAccount
}
