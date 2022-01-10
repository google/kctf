package autoscaling

import (
	kctfv1 "github.com/google/kctf/api/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generate(challenge *kctfv1.Challenge) *autoscalingv1.HorizontalPodAutoscaler {
	// We create the autoscaling object
	autoscaling := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			MaxReplicas:                    challenge.Spec.HorizontalPodAutoscalerSpec.MaxReplicas,
			MinReplicas:                    challenge.Spec.HorizontalPodAutoscalerSpec.MinReplicas,
			TargetCPUUtilizationPercentage: challenge.Spec.HorizontalPodAutoscalerSpec.TargetCPUUtilizationPercentage,
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       challenge.Name,
				APIVersion: "apps/v1",
			},
		},
	}

	return autoscaling
}
