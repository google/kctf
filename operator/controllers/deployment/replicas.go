package deployment

import kctfv1 "github.com/google/kctf/api/v1"

func numReplicas(challenge *kctfv1.Challenge) int32 {
	if challenge.Spec.Deployed == false {
		return 0
	}

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil {
		return -1
	}

	if challenge.Spec.Replicas != nil {
		return *challenge.Spec.Replicas
	}

	return 1
}

func updateNumReplicas(currentReplicas *int32, challenge *kctfv1.Challenge) bool {
	// Updates the number of replicas according to being deployed or not and considering the autoscaling
	replicas := numReplicas(challenge)

	// replicas = -1 means autoscaling is enabled and deployed is true
	if replicas != *currentReplicas && replicas != -1 {
		*currentReplicas = replicas
		return true
	}

	return false
}
