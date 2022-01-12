// Create autoscaling

package autoscaling

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isEqual(autoscalingFound *autoscalingv1.HorizontalPodAutoscaler,
	autoscaling *autoscalingv1.HorizontalPodAutoscaler) bool {
	return reflect.DeepEqual(autoscalingFound.Spec, autoscaling.Spec)
}

func create(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates autoscaling if it doesn't exist yet
	autoscaling := generate(challenge)
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

func delete(autoscalingFound *autoscalingv1.HorizontalPodAutoscaler, client client.Client,
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

func Update(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Creates autoscaling object
	// Checks if an autoscaling was configured
	// If enabled, it checks if it already exists
	autoscalingFound := &autoscalingv1.HorizontalPodAutoscaler{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, autoscalingFound)

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil && errors.IsNotFound(err) &&
		challenge.Spec.Deployed == true {
		// creates autoscaling if it doesn't exist yet
		return create(challenge, client, scheme, log, ctx)
	}

	if (challenge.Spec.HorizontalPodAutoscalerSpec == nil || challenge.Spec.Deployed == false) && err == nil {
		// delete autoscaling
		return delete(autoscalingFound, client, scheme, log, ctx)
	}

	if err == nil {
		if autoscaling := generate(challenge); !isEqual(autoscalingFound, autoscaling) {
			autoscalingFound.Spec = autoscaling.Spec
			err = client.Update(ctx, autoscalingFound)
			if err != nil {
				log.Error(err, "Failed to update autoscaling", "Autoscaling name: ",
					autoscalingFound.Name, " with namespace ", autoscalingFound.Namespace)
				return false, err
			}
			log.Info("Updated autoscaling successfully", "Autoscaling name: ",
				autoscalingFound.Name, " with namespace ", autoscalingFound.Namespace)
			return true, nil
		}
	}

	return false, nil
}
