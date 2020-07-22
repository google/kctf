// File that checks if all configurations are correctly set
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

func CheckDeployed(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// First, check if the challenge is ready and, if not, set replicas to 0
	if challenge.Spec.Deployed == false && *deployment.Spec.Replicas != 0 {
		var numReplicas int32 = 0
		deployment.Spec.Replicas = &numReplicas
		return true
	}

	return false
}

func CheckPowDifficultySeconds() bool {
	// TODO
	return false
}

func CheckNetworkSpecs() bool {
	// Service is created in challenge_controller and here we just check if everything is alright
	// TODO: Do we recheck ports here then?
	// TODO: public
	// TODO: dns
	return false
}

func CheckHealthcheck(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	if challenge.Spec.Healthcheck.Enabled == true {
		// TODO: add other deployment?
	}
	return false
}

// Check autoscaling specs
func CheckAutoscaling(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// Flag to say if there was any change done
	change := false

	if challenge.Spec.Autoscaling.Enabled == true && challenge.Spec.Deployed == true {

		minRep := challenge.Spec.Autoscaling.MinReplicas
		maxRep := challenge.Spec.Autoscaling.MaxReplicas

		if *deployment.Spec.Replicas < minRep {
			deployment.Spec.Replicas = &minRep
			change = true
		}

		if *deployment.Spec.Replicas > maxRep {
			deployment.Spec.Replicas = &maxRep
			change = true
		}

		// TODO: TargetCPUUtilizationPercentage: change resources
	}

	return change
}

func CheckPodTemplate() bool {
	// TODO
	return false
}

func CheckPersistentVolumeClaim() bool {
	// TODO
	return false
}

// TODO: Put reqLogger ?
func CheckConfigurations(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// If any check returns true, it should be requeued;
	requeued := CheckDeployed(challenge, deployment) || CheckPowDifficultySeconds() ||
		CheckNetworkSpecs() || CheckHealthcheck(challenge, deployment) ||
		CheckAutoscaling(challenge, deployment) || CheckPodTemplate() ||
		CheckPersistentVolumeClaim()

	return requeued
}
