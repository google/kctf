// Create autoscaling

package challenge

import (
	"context"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileChallenge) autoscalingForChallenge(challenge *kctfv1alpha1.Challenge) *autoscalingv1.HorizontalPodAutoscaler {
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

func (r *ReconcileChallenge) CreateAutoscaling(challenge *kctfv1alpha1.Challenge,
	ctx context.Context) (reconcile.Result, error) {
	// creates autoscaling if it doesn't exist yet
	autoscaling := r.autoscalingForChallenge(challenge)
	r.log.Info("Creating a Autoscaling")
	err := r.client.Create(ctx, autoscaling)

	if err != nil {
		r.log.Error(err, "Failed to create Autoscaling")
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileChallenge) DeleteAutoscaling(autoscalingFound *autoscalingv1.HorizontalPodAutoscaler,
	ctx context.Context) (reconcile.Result, error) {
	r.log.Info("Deleting Autoscaling")

	err := r.client.Delete(ctx, autoscalingFound)
	if err != nil {
		r.log.Error(err, "Failed to delete Autoscaling")
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}
