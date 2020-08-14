// File that ensures if all configurations are correctly set
package update

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"github.com/google/kctf/pkg/controller/challenge/autoscaling"
	"github.com/google/kctf/pkg/controller/challenge/deployment"
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

// Check if the arrays of ports are the same
func EqualPorts(found []corev1.ServicePort, wanted []corev1.ServicePort) bool {
	if len(found) != len(wanted) {
		return false
	}

	for i, _ := range found {
		if !reflect.DeepEqual(found[i], wanted[i]) {
			return false
		}
	}
	return true
}

func CopyPorts(found *corev1.Service, wanted *corev1.Service) {
	found.Spec.Ports = []corev1.ServicePort{}
	found.Spec.Ports = append(found.Spec.Ports, wanted.Spec.Ports...)
}

func UpdateNumReplicas(challenge *kctfv1alpha1.Challenge, currentReplicas *int32) bool {
	// Updates the number of replicas according to being deployed or not and considering the autoscaling
	var numReplicas int32
	change := false

	if challenge.Spec.Deployed == false && *currentReplicas != 0 {
		numReplicas = 0
		change = true
	}

	if challenge.Spec.Deployed == true && *currentReplicas == 0 &&
		challenge.Spec.HorizontalPodAutoscalerSpec == nil {
		numReplicas = 1
		change = true
	}

	if change == true {
		*currentReplicas = numReplicas
		return true
	}

	return false
}

func UpdateDeployment(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	change := false

	deploymentFound := &appsv1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, deploymentFound)

	if err != nil {
		log.Error(err, "Couldn't get the deployment")
		return false, err
	}

	// Checks if the deployment is correctly set
	if dep := deployment.DeploymentForChallenge(challenge); !reflect.DeepEqual(deploymentFound.Spec.Template.Spec,
		dep.Spec.Template.Spec) {
		change = true
		deploymentFound.Spec.Template.Spec = dep.Spec.Template.Spec
	}

	// Ensure if the challenge is ready and, if not, set replicas to 0
	change = (change || UpdateNumReplicas(challenge, deploymentFound.Spec.Replicas))

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

func UpdatePowDifficultySeconds(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// TODO
	return false, nil
}

func UpdateNetworkSpecs(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
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
	if err != nil && errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		return service.CreateServiceAndIngress(challenge, client, scheme, log, ctx, err_ingress)

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		return service.DeleteServiceAndIngress(serviceFound, ingressFound, client, scheme, log, ctx, err_ingress)
	}

	// Now we check if the service and the ingress are according to the CR:
	if challenge.Spec.Network.Public {
		serv, ingress := service.ServiceForChallenge(challenge)
		// !EqualPorts(serviceFound.Spec.Ports, serv.Spec.Ports)
		// TODO: fix here; we can solve the initial problem by setting the same cluster ip?
		// Comparing specs and doing the same as deployment didn't work;
		if serviceFound.Spec.Ports[0].Port != serviceFound.Spec.Ports[0].Port {
			CopyPorts(serviceFound, serv)
			err = client.Update(ctx, serviceFound)
			if err != nil {
				log.Error(err, "Failed to update service")
				return false, err
			}
			log.Info("Updated service successfully")
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

func UpdatePersistentVolumeClaim(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// TODO
	return false, nil
}

func UpdateAutoscaling(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Creates autoscaling object
	// Checks if an autoscaling was configured
	// If enabled, it checks if it already exists
	autoscalingFound := &autoscalingv1.HorizontalPodAutoscaler{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, autoscalingFound)

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil && err != nil && errors.IsNotFound(err) &&
		challenge.Spec.Deployed == true {
		// creates autoscaling if it doesn't exist yet
		return autoscaling.CreateAutoscaling(challenge, client, scheme, log, ctx)
	}

	if (challenge.Spec.HorizontalPodAutoscalerSpec == nil || challenge.Spec.Deployed == false) && err == nil {
		// delete autoscaling
		return autoscaling.DeleteAutoscaling(autoscalingFound, client, scheme, log, ctx)
	}

	if err == nil {
		if autoscaling := autoscaling.AutoscalingForChallenge(challenge); !reflect.DeepEqual(autoscalingFound.Spec,
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

func UpdateConfigurations(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// We check if there's an error in each update
	updateFunctions := []func(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
		log logr.Logger, ctx context.Context) (bool, error){UpdateNetworkSpecs, UpdateAutoscaling, UpdateDeployment,
		UpdatePersistentVolumeClaim, UpdatePowDifficultySeconds}

	for _, updateFunction := range updateFunctions {
		requeue, err := updateFunction(challenge, cl, scheme, log, ctx)
		if err != nil {
			return requeue, err
		}
	}

	return false, nil
}
