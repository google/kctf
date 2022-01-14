package resources

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewNetworkPolicyBlockInternal() client.Object {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "block-internal",
			Namespace: "default",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				"Egress",
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{{
				To: []networkingv1.NetworkPolicyPeer{{
					IPBlock: &networkingv1.IPBlock{
						CIDR: "0.0.0.0/0",
						Except: []string{"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10",
							"127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24",
							"192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15",
							"198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4"},
					},
				}},
			}},
		},
	}
	return networkPolicy
}
