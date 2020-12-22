package dns

import (
	netgkev1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generate(domainName string, challenge *kctfv1alpha1.Challenge) *netgkev1.ManagedCertificate {
	certificate := &netgkev1.ManagedCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: netgkev1.ManagedCertificateSpec{
			Domains: []string{
				challenge.Name + "-http." + domainName,
			},
		},
		Status: netgkev1.ManagedCertificateStatus{
			CertificateStatus: "",
			DomainStatus:      []netgkev1.DomainStatus{},
			CertificateName:   "",
			ExpireTime:        "",
		},
	}

	return certificate
}
