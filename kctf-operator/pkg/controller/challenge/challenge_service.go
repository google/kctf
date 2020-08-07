// Creates the service

package challenge

import (
	"context"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileChallenge) serviceForChallenge(challenge *kctfv1alpha1.Challenge) (*corev1.Service, *netv1beta1.Ingress) {
	// Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: "LoadBalancer",
		},
	}

	// Ingress object
	ingress := &netv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "https",
			Namespace:   challenge.Namespace,
			Labels:      map[string]string{"app": challenge.Name},
			Annotations: map[string]string{"networking.gke.io/managed-certificates": "kctf-certificate"},
		},
	}

	for _, port := range challenge.Spec.Network.Ports {
		if port.Protocol == "HTTPS" {
			// If not declared
			if port.Port == 0 {
				port.Port = 1
			}

			// Creates the ingress object
			ingress.Spec.Backend = &netv1beta1.IngressBackend{
				ServiceName: challenge.Name,
				ServicePort: intstr.FromInt(int(port.Port)),
			}

			servicePort := corev1.ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort,
			}
			service.Spec.Ports = append(service.Spec.Ports, servicePort)
		} else {
			// Creates the port
			servicePort := corev1.ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort,
				Protocol:   port.Protocol,
			}
			service.Spec.Ports = append(service.Spec.Ports, servicePort)
		}
	}

	return service, ingress
}

func (r *ReconcileChallenge) CreateServiceAndIngress(challenge *kctfv1alpha1.Challenge, ctx context.Context,
	err_ingress error) (reconcile.Result, error) {
	serv, ingress := r.serviceForChallenge(challenge)
	// See if there's any port defined for the service
	r.log.Info("Creating a new Service", "Service.Namespace",
		serv.Namespace, "Service.Name", serv.Name)
	err := r.client.Create(ctx, serv)

	if err != nil {
		r.log.Error(err, "Failed to create new Service", "Service.Namespace",
			serv.Namespace, "Service.Name", serv.Name)
		return reconcile.Result{}, err
	}

	// Create ingress, if there's any https
	if err_ingress != nil && errors.IsNotFound(err_ingress) {
		// If there's a port HTTPS
		if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == true {
			// Create ingress in the client
			r.log.Info("Creating a new Ingress", "Ingress.Namespace", ingress.Namespace,
				"Ingress.Name", ingress.Name)
			err = r.client.Create(ctx, ingress)

			if err != nil {
				r.log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", ingress.Namespace,
					"Ingress.Name", ingress.Name)
				return reconcile.Result{}, err
			}

			// Ingress created successfully
			return reconcile.Result{}, err
		}

		if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == false {
			r.log.Info("Failed to create Ingress instance, DNS isn't enabled. Challenge won't be reconciled here.")
		}
	}

	// Service created successfully - return and requeue
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileChallenge) DeleteServiceAndIngress(serviceFound *corev1.Service, ingressFound *netv1beta1.Ingress,
	ctx context.Context, err_ingress error) (reconcile.Result, error) {
	r.log.Info("Deleting the Service", "Service.Namespace", serviceFound.Namespace,
		"Service.Name", serviceFound.Name)
	err := r.client.Delete(ctx, serviceFound)

	if err != nil {
		r.log.Error(err, "Failed to delete Service", "Service.Namespace", serviceFound.Namespace,
			"Service.Name", serviceFound.Name)
		return reconcile.Result{}, err
	}

	// Delete ingress if existent
	if err_ingress == nil {
		r.log.Info("Deleting the Ingress", "Ingress.Namespace", ingressFound.Namespace, "Ingress.Name", ingressFound.Name)
		err = r.client.Delete(ctx, ingressFound)

		if err != nil {
			r.log.Error(err, "Failed to delete Ingress", "Ingress.Namespace", ingressFound.Namespace,
				"Ingress.Name", ingressFound.Name)
			return reconcile.Result{}, err
		}
	}

	// Service deleted successfully - return and requeue
	return reconcile.Result{}, err
}
