// File that ensures if all configurations are correctly set
// TODO: create errors in case we can't get the instance (error different from not found)
package update

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"github.com/google/kctf/pkg/controller/challenge/autoscaling"
	"github.com/google/kctf/pkg/controller/challenge/deployment"
	"github.com/google/kctf/pkg/controller/challenge/service"
	"github.com/google/kctf/pkg/controller/challenge/volumes"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func updateDeployment(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Flags if there was a change
	change := false

	deploymentFound := &appsv1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, deploymentFound)

	if err != nil {
		log.Error(err, "Couldn't get the deployment")
		return false, err
	}

	// Checks if the deployment is correctly set
	if dep := deployment.Generate(challenge); !reflect.DeepEqual(deploymentFound.Spec.Template.Spec,
		dep.Spec.Template.Spec) {
		change = true
		deploymentFound.Spec.Template.Spec = dep.Spec.Template.Spec
	}

	// Ensure if the challenge is ready and, if not, set replicas to 0
	change = (change || updateNumReplicas(challenge, deploymentFound.Spec.Replicas))

	// Updates deployment with client
	if change == true {
		err = client.Update(ctx, deploymentFound)
		if err != nil {
			log.Error(err, "Failed to update deployment")
			return false, err
		}
		log.Info("Deployment updated succesfully")
		return true, nil
	}

	return false, nil
}

func updatePowDifficultySeconds(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// TODO: create configmap and apply secrets
	return false, nil
}

func updateNetworkSpecs(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
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
	if errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		return service.Create(challenge, client, scheme, log, ctx, err_ingress)

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		return service.Delete(serviceFound, ingressFound, client, scheme, log, ctx, err_ingress)
	}

	// Now we check if the service and the ingress are according to the CR:
	if challenge.Spec.Network.Public {
		serv, ingress := service.Generate(challenge)
		if !equalPorts(serviceFound.Spec.Ports, serv.Spec.Ports) {
			copyPorts(serviceFound, serv)
			err = client.Update(ctx, serviceFound)
			if err != nil {
				log.Error(err, "Failed to update service")
				return false, err
			}
			log.Info("Service updated successfully")
			return true, nil
		}
		// Flags if there was a change in the ingress instance
		change_ingress := false

		// TODO: check dns and domain name here

		// If ingress should be created:
		if errors.IsNotFound(err_ingress) && ingress.Spec.Backend != nil {
			// create ingress
			change_ingress = true
			err = client.Create(ctx, ingress)
		}

		// Cases when the ingress should be deleted or merely updated
		if err_ingress == nil && !reflect.DeepEqual(ingressFound.Spec, ingress.Spec) {
			change_ingress = true
			if ingressFound.Spec.Backend != nil && ingress.Spec.Backend == nil {
				// Deletes ingress
				err = client.Delete(ctx, ingressFound)
			} else {
				// Updates ingress
				ingressFound.Spec = ingress.Spec
				err = client.Update(ctx, ingressFound)
			}
		}

		if change_ingress == true {
			if err != nil {
				log.Error(err, "Failed to update ingress")
				return false, err
			}
			log.Info("Updated ingress successfully")
			return true, nil
		}
	}

	return false, nil
}

// TODO: should the whole block be done at once and check the error only in the end
// or should we check in each interation with the client
func updatePersistentVolumeClaims(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Check if all persistent volume claims are correctly set and update them if necessary
	// We get all persistentVolumeClaims in the same namespace as the challenge
	persistentVolumeClaimsFound := &corev1.PersistentVolumeClaimList{}
	change := false

	// List all persistent volume claims in the namespace of the challenge
	var listOption client.ListOption
	listOption = &client.ListOptions{
		Namespace: challenge.Namespace,
	}

	err := cl.List(ctx, persistentVolumeClaimsFound, listOption)
	if err != nil {
		log.Error(err, "Failed to list persistent volume claims")
		return false, err
	}

	// First we create a map with the names of the persistent volume claims that already exist
	namesFound := mapNameIdx(persistentVolumeClaimsFound)

	// For comparing two persistentVolumeClaims, we will use DeepEqual
	if challenge.Spec.PersistentVolumeClaims != nil {
		for i, claim := range challenge.Spec.Claims {
			value, present := namesFound[claim]
			if present == true {
				delete(namesFound, item.Name)
			} else {
				// Creates the object
				change, err = volumes.Create(challenge, claim,
					cl, scheme, log, ctx)
				if err != nil {
					return false, err
				}
			}
		}
	}

	// Then we delete the persistent volume claims that remained
	for _, idx := range namesFound {
		change, err = volumes.Delete(&persistentVolumeClaimsFound.Items[idx],
			cl, scheme, log, ctx)
		if err != nil {
			return false, err
		}
	}

	return change, err
}

func updateAutoscaling(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
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
		return autoscaling.Create(challenge, client, scheme, log, ctx)
	}

	if (challenge.Spec.HorizontalPodAutoscalerSpec == nil || challenge.Spec.Deployed == false) && err == nil {
		// delete autoscaling
		return autoscaling.Delete(autoscalingFound, client, scheme, log, ctx)
	}

	if err == nil {
		if autoscaling := autoscaling.Generate(challenge); !reflect.DeepEqual(autoscalingFound.Spec,
			autoscaling.Spec) {
			autoscalingFound.Spec = autoscaling.Spec
			err = client.Update(ctx, autoscalingFound)
			if err != nil {
				log.Error(err, "Failed to update autoscaling")
				return false, err
			}
			log.Info("Updated autoscaling successfully")
			return true, nil
		}
	}

	return false, nil
}

func Configurations(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// We check if there's an error in each update
	updateFunctions := []func(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
		log logr.Logger, ctx context.Context) (bool, error){updatePersistentVolumeClaims,
		updatePowDifficultySeconds, updateDeployment, updateNetworkSpecs, updateAutoscaling}

	for _, updateFunction := range updateFunctions {
		requeue, err := updateFunction(challenge, cl, scheme, log, ctx)
		if err != nil {
			return requeue, err
		}
	}

	return false, nil
}
