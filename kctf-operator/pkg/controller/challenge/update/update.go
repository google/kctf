// File that ensures if all configurations are correctly set
package update

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"github.com/google/kctf/pkg/controller/challenge/autoscaling"
	"github.com/google/kctf/pkg/controller/challenge/service"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateDeployment(challenge *kctfv1alpha1.Challenge, deployment *appsv1.Deployment) bool {
	// First, ensure if the challenge is ready and, if not, set replicas to 0
	// TODO: inline this, reference:
	// TODO: Check if autoscaling is enabled
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

func UpdateNetworkSpecs(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// TODO: Do we check ports here then?
	// TODO: dns
	// Service is created in challenge_controller and here we just ensure that everything is alright
	// Creates the service if it doesn't exist
	// Check existence of the service:
	serviceFound := &corev1.Service{}
	ingressFound := &netv1beta1.Ingress{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, serviceFound)
	err_ingress := client.Get(ctx, types.NamespacedName{Name: "https",
		Namespace: challenge.Namespace}, ingressFound)

	// Just enter here if the service doesn't exist yet:
	if err != nil && errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		return service.CreateServiceAndIngress(challenge, client, scheme, log, ctx, err_ingress)

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		return service.DeleteServiceAndIngress(serviceFound, ingressFound, client, scheme, log, ctx, err_ingress)
	}
	return false, nil
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

func UpdateAutoscaling(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Creates autoscaling object
	// Checks if an autoscaling was configured
	// If enabled, it checks if it already exists
	autoscalingFound := &autoscalingv1.HorizontalPodAutoscaler{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, autoscalingFound)

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil && err != nil && errors.IsNotFound(err) {
		// creates autoscaling if it doesn't exist yet
		return autoscaling.CreateAutoscaling(challenge, client, scheme, log, ctx)
	}

	if challenge.Spec.HorizontalPodAutoscalerSpec == nil && err == nil {
		// delete autoscaling
		return autoscaling.DeleteAutoscaling(autoscalingFound, client, scheme, log, ctx)
	}

	return false, nil
}

func UpdateConfigurations(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// We check if there's an error in each update
	updateFunctions := []func(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
		log logr.Logger, ctx context.Context) (bool, error){UpdateNetworkSpecs, UpdateAutoscaling}

	for _, updateFunction := range updateFunctions {
		requeue, err := updateFunction(challenge, cl, scheme, log, ctx)
		if err != nil {
			return requeue, err
		}
	}

	return false, nil
}
