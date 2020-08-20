// Create autoscaling

package autoscaling

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Create(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates autoscaling if it doesn't exist yet
	autoscaling := Generate(challenge)
	log.Info("Creating a Autoscaling", "Autoscaling name: ",
		autoscaling.Name, " with namespace ", autoscaling.Namespace)

	// Creates owner references
	err := controllerutil.SetControllerReference(challenge, autoscaling, scheme)

	// Creates autoscaling
	err = client.Create(ctx, autoscaling)

	if err != nil {
		log.Error(err, "Failed to create Autoscaling", "Autoscaling name: ",
			autoscaling.Name, " with namespace ", autoscaling.Namespace)
		return false, err
	}

	return true, nil
}

func Delete(autoscalingFound *autoscalingv1.HorizontalPodAutoscaler, client client.Client,
	scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	log.Info("Deleting Autoscaling", "Autoscaling name: ",
		autoscalingFound.Name, " with namespace ", autoscalingFound.Namespace)

	err := client.Delete(ctx, autoscalingFound)
	if err != nil {
		log.Error(err, "Failed to delete Autoscaling", "Autoscaling name: ",
			autoscalingFound.Name, " with namespace ", autoscalingFound.Namespace)
		return false, err
	}

	return true, nil
}
