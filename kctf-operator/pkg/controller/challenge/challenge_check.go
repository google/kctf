// File that checks if all configurations are correctly set
package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// TODO: Put reqLogger ?
func CheckConfigurations(challenge *kctfv1alpha1.Challenge, found *appsv1.Deployment) bool {

	change := false
	if challenge.Spec.Deployed == false {
		var numReplicas int32 = 0
		found.Spec.Replicas = &numReplicas
		change = true
	}

	// Check powDifficultySeconds

	// TODO: how do we set powDifficulty ?

	// Check network specs:

	// TODO: public
	// TODO: dns

	// Ports are set in the deployment

	// Check healthcheck specs

	if challenge.Spec.Healthcheck.Enabled == true {
		// TODO: add other deployment?
	}

	// Check autoscaling specs

	if challenge.Spec.Autoscaling.Enabled == true && challenge.Spec.Deployed == true {
		minRep := challenge.Spec.Autoscaling.MinReplicas
		maxRep := challenge.Spec.Autoscaling.MaxReplicas

		if *found.Spec.Replicas < minRep {
			found.Spec.Replicas = &minRep
			change = true
		}

		if *found.Spec.Replicas > maxRep {
			found.Spec.Replicas = &maxRep
			change = true
		}

		// TODO: TargetCPUUtilizationPercentage: change resources
	}

	// Check deployment specs

	if challenge.Spec.Deployment.Enabled == true {
		// TODO
	}

	// TODO: Check persistentVolumeClaim

	return change
}
