// All the functions that check if there was a change in the object
package update

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Check if the arrays of ports are the same
func equalPorts(found []corev1.ServicePort, wanted []corev1.ServicePort) bool {
	if len(found) != len(wanted) {
		return false
	}

	for i, _ := range found {
		if found[i].Name != wanted[i].Name || found[i].Protocol != wanted[i].Protocol ||
			found[i].Port != wanted[i].Port || found[i].TargetPort != wanted[i].TargetPort {
			return false
		}
	}
	return true
}

// Copy ports from one service to another
func copyPorts(found *corev1.Service, wanted *corev1.Service) {
	found.Spec.Ports = []corev1.ServicePort{}
	found.Spec.Ports = append(found.Spec.Ports, wanted.Spec.Ports...)
}

func mapNameIdx(persistentVolumeClaimsFound *corev1.PersistentVolumeClaimList) map[string]int {
	m := make(map[string]int)

	for idx, item := range persistentVolumeClaimsFound.Items {
		m[item.Name] = idx
	}

	return m
}

func updateNumReplicas(challenge *kctfv1alpha1.Challenge, currentReplicas *int32) bool {
	// Updates the number of replicas according to being deployed or not and considering the autoscaling
	var numReplicas int32
	change := false

	// TODO: Inline this?
	if challenge.Spec.Deployed == false && *currentReplicas != 0 {
		numReplicas = 0
		change = true
	}

	if challenge.Spec.Deployed == true && *currentReplicas == 0 &&
		challenge.Spec.HorizontalPodAutoscalerSpec == nil {
		numReplicas = 1
		change = true
	}

	if change == true {
		*currentReplicas = numReplicas
		return true
	}

	return false
}
