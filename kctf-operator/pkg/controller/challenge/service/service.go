package service

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func Generate(challenge *kctfv1alpha1.Challenge) (*corev1.Service, *netv1beta1.Ingress) {
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
				Protocol:   "TCP",
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
