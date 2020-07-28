// Create autoscaling

package challenge

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ReconcileChallenge) autoscalingForChallenge(challenge *kctfv1alpha1.Challenge) *autoscalingv1.HorizontalPodAutoscaler {
	// First we change the Spec of the autoscaling, so it targets the deployment
	challenge.Spec.HorizontalAutoscaling.ScaleTargetRef = autoscalingv1.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       challenge.Name,
		APIVersion: "apps/v1",
	}

	// Then, we create the autoscaling object
	autoscaling := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: challenge.Spec.HorizontalAutoscaling,
	}

	return autoscaling
}
