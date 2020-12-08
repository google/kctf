package network

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func generatePolicies(challenge *kctfv1alpha1.Challenge) []netv1.NetworkPolicy {
	blockInternalPolicy := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "block-internal",
			Namespace: challenge.Namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PolicyTypes: []netv1.PolicyType{"Egress"},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "0.0.0.0/0",
								Except: []string{
									"0.0.0.0/8",
									"10.0.0.0/8",
									"100.64.0.0/10",
									"127.0.0.0/8",
									"169.254.0.0/16",
									"172.16.0.0/12",
									"192.0.0.0/24",
									"192.0.2.0/24",
									"192.88.99.0/24",
									"192.168.0.0/16",
									"198.18.0.0/15",
									"198.51.100.0/24",
									"203.0.113.0/24",
									"224.0.0.0/4",
									"240.0.0.0/4",
								},
							},
						},
					},
				},
			},
		},
	}

	udp := v1.ProtocolUDP
	tcp := v1.ProtocolTCP
	port53 := intstr.FromInt(53)

	allowDNSPolicy := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-dns",
			Namespace: challenge.Namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PolicyTypes: []netv1.PolicyType{"Egress"},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							Protocol: &udp,
							Port:     &port53,
						},
						{
							Protocol: &tcp,
							Port:     &port53,
						},
					},
				},
			},
		},
	}

	return []netv1.NetworkPolicy{blockInternalPolicy, allowDNSPolicy}
}
