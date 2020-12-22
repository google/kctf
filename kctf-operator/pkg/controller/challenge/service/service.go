package service

import (
	"strconv"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func generateClusterIPService(challenge *kctfv1alpha1.Challenge) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
			Labels:    map[string]string{"app": challenge.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": challenge.Name},
			Type:     "ClusterIP",
			Ports:    []corev1.ServicePort{},
		},
	}
	for _, port := range challenge.Spec.Network.Ports {
		protocol := corev1.ProtocolTCP
		switch port.Protocol {
		case corev1.ProtocolSCTP, corev1.ProtocolTCP, corev1.ProtocolUDP:
			protocol = port.Protocol
		}
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Port:       port.TargetPort.IntVal,
			TargetPort: port.TargetPort,
			Protocol:   protocol,
		})
	}

	return service
}

func generateLoadBalancerService(domainName string, challenge *kctfv1alpha1.Challenge) (*corev1.Service, *netv1beta1.Ingress) {
	// Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name + "-lb-service",
			Namespace: challenge.Namespace,
			Labels:    map[string]string{"app": challenge.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": challenge.Name},
			Type:     "LoadBalancer",
		},
	}

	// Ingress object
	ingress := &netv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        challenge.Name,
			Namespace:   challenge.Namespace,
			Labels:      map[string]string{"app": challenge.Name},
			Annotations: map[string]string{"networking.gke.io/managed-certificates": challenge.Name},
		},
		Spec: netv1beta1.IngressSpec{
			Rules: []netv1beta1.IngressRule{{
				Host: challenge.Name + "-http." + domainName,
			}},
		},
	}

	for i, port := range challenge.Spec.Network.Ports {
		if port.Protocol == "HTTPS" {
			// If not declared
			if port.Port == 0 {
				port.Port = 1
			}

			// Creates the ingress object
			ingress.Spec.Backend = &netv1beta1.IngressBackend{
				ServiceName: service.Name,
				ServicePort: intstr.FromInt(int(port.Port)),
			}

			servicePort := corev1.ServicePort{
				Port:       port.Port,
				TargetPort: port.TargetPort,
				Protocol:   "TCP",
			}

			if port.Name != "" {
				servicePort.Name = port.Name
			} else {
				servicePort.Name = "port-" + strconv.Itoa(i)
			}

			service.Spec.Ports = append(service.Spec.Ports, servicePort)
		} else {
			// Creates the port
			servicePort := corev1.ServicePort{
				Port:       port.Port,
				TargetPort: port.TargetPort,
				Protocol:   port.Protocol,
			}

			if port.Name != "" {
				servicePort.Name = port.Name
			} else {
				servicePort.Name = "port-" + strconv.Itoa(i)
			}

			service.Spec.Ports = append(service.Spec.Ports, servicePort)
		}
	}

	// Add annotation in the case it's a web challenge
	if ingress.Spec.Backend != nil && domainName != "" &&
		challenge.Spec.Network.Dns == true {
		service.ObjectMeta.Annotations =
			map[string]string{"external-dns.alpha.kubernetes.io/hostname": challenge.Name +
				"-tcp." + domainName}
	}
	return service, ingress
}
