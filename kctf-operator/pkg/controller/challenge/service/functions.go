// Creates the service

package service

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	utils "github.com/google/kctf/pkg/controller/challenge/utils"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isServiceEqual(serviceFound *corev1.Service, serv *corev1.Service) bool {
	return equalPorts(serviceFound.Spec.Ports, serv.Spec.Ports)
}

func isIngressEqual(ingressFound *netv1beta1.Ingress, ingress *netv1beta1.Ingress) bool {
	return reflect.DeepEqual(ingressFound.Spec, ingress.Spec)
}

// Check if the arrays of ports are the same
func equalPorts(found []corev1.ServicePort, wanted []corev1.ServicePort) bool {
	if len(found) != len(wanted) {
		return false
	}

	for i, _ := range found {
		if found[i].Name != wanted[i].Name || found[i].Protocol != wanted[i].Protocol ||
			found[i].Port != wanted[i].Port || found[i].TargetPort != wanted[i].TargetPort {
			return false
		}
	}
	return true
}

// Copy ports from one service to another
func copyPorts(found *corev1.Service, wanted *corev1.Service) {
	found.Spec.Ports = []corev1.ServicePort{}
	found.Spec.Ports = append(found.Spec.Ports, wanted.Spec.Ports...)
}

func create(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context, errIngress error) (bool, error) {
	domainName := utils.GetDomainName(challenge, client, log, ctx)
	serv, ingress := generateLoadBalancerService(domainName, challenge)
	// Create the service
	log.Info("Creating a new Service", "Service.Namespace",
		serv.Namespace, "Service.Name", serv.Name)

	// Defines ownership
	controllerutil.SetControllerReference(challenge, serv, scheme)

	// Creates the service
	err := client.Create(ctx, serv)

	if err != nil {
		log.Error(err, "Failed to create new Service", "Service.Namespace",
			serv.Namespace, "Service.Name", serv.Name)
		return false, err
	}

	// Create ingress, if there's any https
	if errors.IsNotFound(errIngress) && ingress.Spec.Backend != nil {
		// If there's a port HTTPS
		if challenge.Spec.Network.Dns == true && domainName != "" {
			// Create ingress in the client
			log.Info("Creating a new Ingress", "Ingress Namespace", ingress.Namespace,
				"Ingress Name", ingress.Name)

			// Defines ownership
			controllerutil.SetControllerReference(challenge, ingress, scheme)

			// Creates the ingress
			err = client.Create(ctx, ingress)

			if err != nil {
				log.Error(err, "Failed to create new Ingress", "Ingress Namespace", ingress.Namespace,
					"Ingress Name", ingress.Name)
				return false, err
			}
		} else {
			// If there was some inconsistency with dns or domain name
			if challenge.Spec.Network.Dns == false {
				log.Info("Failed to create Ingress instance, DNS isn't enabled. Challenge won't be reconciled here.",
					"Challenge name: ", challenge.Name, " with namespace ", challenge.Namespace)
			}

			if domainName == "" {
				log.Info("Failed to create Ingress instance, domain name wasn't set. Challenge won't be reconciled here.",
					"Challenge name: ", challenge.Name, " with namespace ", challenge.Namespace)
			}

			return false, nil
		}
	}

	// Service created successfully - return and requeue
	return true, nil
}

func delete(serviceFound *corev1.Service, ingressFound *netv1beta1.Ingress,
	client client.Client, scheme *runtime.Scheme, log logr.Logger,
	ctx context.Context, errIngress error) (bool, error) {
	log.Info("Deleting the Service", "Service.Namespace", serviceFound.Namespace,
		"Service.Name", serviceFound.Name)
	err := client.Delete(ctx, serviceFound)

	if err != nil {
		log.Error(err, "Failed to delete Service", "Service.Namespace", serviceFound.Namespace,
			"Service.Name", serviceFound.Name)
		return false, err
	}

	// Delete ingress if existent
	if errIngress == nil {
		log.Info("Deleting the Ingress", "Ingress.Namespace", ingressFound.Namespace, "Ingress.Name", ingressFound.Name)
		err = client.Delete(ctx, ingressFound)

		if err != nil {
			log.Error(err, "Failed to delete Ingress", "Ingress.Namespace", ingressFound.Namespace,
				"Ingress.Name", ingressFound.Name)
			return false, err
		}
	}

	// Service deleted successfully - return and requeue
	return true, err
}

func updateInternalService(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	newService := generateClusterIPService(challenge)
	existingService := &corev1.Service{}
	err := client.Get(ctx, types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, existingService)

	if err == nil {
		// client.Get successful: try to update the existing service
		if isServiceEqual(existingService, newService) {
			return false, nil
		}

		copyPorts(existingService, newService)
		err = client.Update(ctx, existingService)
		if err != nil {
			return false, err
		}

		log.Info("Updated internal service successfully", " Name: ",
			newService.Name, " with namespace ", newService.Namespace)
		return true, nil
	}

	if !errors.IsNotFound(err) {
		return false, err
	}

	// Defines ownership
	controllerutil.SetControllerReference(challenge, newService, scheme)

	// Creates the service
	err = client.Create(ctx, newService)
	if err != nil {
		return false, err
	}

	log.Info("Created internal service successfully", " Name: ",
		newService.Name, " with namespace ", newService.Namespace)

	return true, nil
}

func updateLoadBalancerService(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Service is created in challenge_controller and here we just ensure that everything is alright
	// Creates the service if it doesn't exist
	// Check existence of the service:
	serviceFound := &corev1.Service{}
	ingressFound := &netv1beta1.Ingress{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name + "-lb-service",
		Namespace: challenge.Namespace}, serviceFound)
	errIngress := client.Get(ctx, types.NamespacedName{Name: "https",
		Namespace: challenge.Namespace}, ingressFound)

	// Get the domainName
	domainName := utils.GetDomainName(challenge, client, log, ctx)

	// Just enter here if the service doesn't exist yet:
	if errors.IsNotFound(err) && challenge.Spec.Network.Public == true &&
		challenge.Spec.Deployed == true {
		// Define a new service if the challenge is public
		return create(challenge, client, scheme, log, ctx, errIngress)

		// When service exists and public is changed to false
	} else if err == nil && (challenge.Spec.Network.Public == false ||
		challenge.Spec.Deployed == false) {
		return delete(serviceFound, ingressFound, client, scheme, log, ctx, errIngress)
	}

	// Now we check if the service and the ingress are according to the CR:
	if challenge.Spec.Network.Public && challenge.Spec.Deployed == true {
		serv, ingress := generateLoadBalancerService(domainName, challenge)
		if !isServiceEqual(serviceFound, serv) {
			copyPorts(serviceFound, serv)
			err = client.Update(ctx, serviceFound)
			if err != nil {
				log.Error(err, "Failed to update service", "Service Name: ",
					serv.Name, " with namespace ", serv.Namespace)
				return false, err
			}
			log.Info("Service updated successfully", "Name: ",
				serv.Name, " with namespace ", serv.Namespace)
			return true, nil
		}
		// Flags if there was a change in the ingress instance
		change_ingress := false

		// If ingress should be created:
		if errors.IsNotFound(errIngress) && ingress.Spec.Backend != nil {
			// create ingress
			change_ingress = true
			err = client.Create(ctx, ingress)
		}

		// Cases when the ingress should be deleted or merely updated
		if errIngress == nil && !isIngressEqual(ingressFound, ingress) {
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
				log.Error(err, "Failed to update ingress", "Ingress Name: ",
					ingress.Name, " with namespace ", ingress.Namespace)
				return false, err
			}
			log.Info("Updated ingress successfully", "Ingress Name: ",
				ingress.Name, " with namespace ", ingress.Namespace)
			return true, nil
		}
	}

	return false, nil
}

func Update(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	changed := false

	internalServiceChanged, err := updateInternalService(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating internal service", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || internalServiceChanged

	loadBalancerServiceChanged, err := updateLoadBalancerService(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating load balancer service", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace, " error: ")
		return false, err
	}
	changed = changed || loadBalancerServiceChanged

	return changed, nil
}
