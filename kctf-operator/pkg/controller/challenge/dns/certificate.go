package dns

import (
	netgkev1beta1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1beta1"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generate(domainName string, challenge *kctfv1alpha1.Challenge) *netgkev1beta1.ManagedCertificate {
	certificate := &netgkev1beta1.ManagedCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kctf-certificate",
			Namespace: challenge.Namespace,
		},
		Spec: netgkev1beta1.ManagedCertificateSpec{
			Domains: []string{
				challenge.Name + "." + domainName,
			},
		},
	}

	return certificate
}
