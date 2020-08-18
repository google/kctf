// Creates the service

package service

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Create(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context, err_ingress error) (bool, error) {
	serv, ingress := Generate(challenge)
	// Create the service
	log.Info("Creating a new Service", "Service.Namespace",
		serv.Namespace, "Service.Name", serv.Name)

	// Defines ownership
	err := controllerutil.SetControllerReference(challenge, serv, scheme)

	// Creates the service
	err = client.Create(ctx, serv)

	if err != nil {
		log.Error(err, "Failed to create new Service", "Service.Namespace",
			serv.Namespace, "Service.Name", serv.Name)
		return false, err
	}

	// Create ingress, if there's any https
	if err_ingress != nil && errors.IsNotFound(err_ingress) {
		// If there's a port HTTPS
		if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == true {
			// Create ingress in the client
			log.Info("Creating a new Ingress", "Ingress.Namespace", ingress.Namespace,
				"Ingress.Name", ingress.Name)

			// Defines ownership
			err := controllerutil.SetControllerReference(challenge, ingress, scheme)

			// Creates the ingress
			err = client.Create(ctx, ingress)

			if err != nil {
				log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", ingress.Namespace,
					"Ingress.Name", ingress.Name)
				return false, err
			}
		}

		if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == false {
			log.Info("Failed to create Ingress instance, DNS isn't enabled. Challenge won't be reconciled here.")
		}
	}

	// Service created successfully - return and requeue
	return true, nil
}

func Delete(serviceFound *corev1.Service, ingressFound *netv1beta1.Ingress,
	client client.Client, scheme *runtime.Scheme, log logr.Logger,
	ctx context.Context, err_ingress error) (bool, error) {
	log.Info("Deleting the Service", "Service.Namespace", serviceFound.Namespace,
		"Service.Name", serviceFound.Name)
	err := client.Delete(ctx, serviceFound)

	if err != nil {
		log.Error(err, "Failed to delete Service", "Service.Namespace", serviceFound.Namespace,
			"Service.Name", serviceFound.Name)
		return false, err
	}

	// Delete ingress if existent
	if err_ingress == nil {
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
