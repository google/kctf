// File that ensures if all configurations are correctly set
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

func UpdateDeployed(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// First, ensure if the challenge is ready and, if not, set replicas to 0
	// TODO: check if horizontal autoscaling is enabled
	if challenge.Spec.Deployed == false && *deployment.Spec.Replicas != 0 {
		var numReplicas int32 = 0
		deployment.Spec.Replicas = &numReplicas
		return true
	}

	if challenge.Spec.Deployed == true && *deployment.Spec.Replicas == 0 {
		var numReplicas int32 = 1
		deployment.Spec.Replicas = &numReplicas
		return true
	}

	return false
}

func UpdatePowDifficultySeconds() bool {
	// TODO
	return false
}

func UpdateNetworkSpecs() bool {
	// Service is created in challenge_controller and here we just ensure that everything is alright
	// TODO: Do we check ports here then?
	// TODO: dns
	return false
}

func UpdateHealthcheck(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	if challenge.Spec.Healthcheck.Enabled == true {
		// TODO
	}
	return false
}

func UpdatePodTemplate() bool {
	// TODO
	return false
}

func UpdatePersistentVolumeClaim() bool {
	// TODO
	return false
}

// TODO: Put reqLogger ?
func UpdateConfigurations(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// If any of the ensures returns true, it should be requeued;
	requeued := UpdateDeployed(challenge, deployment) || UpdatePowDifficultySeconds() ||
		UpdateNetworkSpecs() || UpdateHealthcheck(challenge, deployment) ||
		UpdatePodTemplate() || UpdatePersistentVolumeClaim()

	return requeued
}
