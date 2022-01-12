package network

import (
	"fmt"

	kctfv1 "github.com/google/kctf/api/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generatePolicies(challenge *kctfv1.Challenge) []netv1.NetworkPolicy {
	var egressRules = make([]netv1.NetworkPolicyEgressRule, len(challenge.Spec.AllowConnectTo))
	for i, targetName := range challenge.Spec.AllowConnectTo {
		egressRules[i] = netv1.NetworkPolicyEgressRule{
			To: []netv1.NetworkPolicyPeer{
				{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": targetName,
						},
					},
				},
			},
		}
	}

	challengeAccessPolicy := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-challenge-access", challenge.Name),
			Namespace: challenge.Namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PolicyTypes: []netv1.PolicyType{"Egress"},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": challenge.Name,
				},
			},
			Egress: egressRules,
		},
	}

	return []netv1.NetworkPolicy{challengeAccessPolicy}
}
