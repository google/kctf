package resources

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewAllowDns() client.Object {
	udpProtocol := corev1.ProtocolUDP
	udpPort := intstr.FromInt(53)
	tcpProtocol := corev1.ProtocolTCP
	tcpPort := intstr.FromInt(53)

	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-dns",
			Namespace: "default",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				"Egress",
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{{
				To: []networkingv1.NetworkPolicyPeer{},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: &udpProtocol,
						Port:     &udpPort,
					},
					{
						Protocol: &tcpProtocol,
						Port:     &tcpPort,
					},
				},
			}},
		},
	}

	return networkPolicy
}
