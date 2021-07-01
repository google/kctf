package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	gkenetv1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1"
	backendv1 "github.com/google/kctf/pkg/apis/cloud/v1"
	kctfv1 "github.com/google/kctf/pkg/apis/kctf/v1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func generateNodePortService(challenge *kctfv1.Challenge) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        challenge.Name,
			Namespace:   challenge.Namespace,
			Labels:      map[string]string{"app": challenge.Name},
			Annotations: map[string]string{"cloud.google.com/backend-config": fmt.Sprintf("{\"default\": \"%s\"}", challenge.Name)},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": challenge.Name},
			Type:     "NodePort",
			Ports:    []corev1.ServicePort{},
		},
	}

	portsSeen := make(map[int32]bool)

	for i, port := range challenge.Spec.Network.Ports {
		if portsSeen[port.Port] {
			continue
		}
		portsSeen[port.Port] = true

		protocol := corev1.ProtocolTCP
		switch port.Protocol {
		case corev1.ProtocolSCTP, corev1.ProtocolTCP, corev1.ProtocolUDP:
			protocol = port.Protocol
		}

		servicePort := port.Port
		if servicePort == 0 {
			servicePort = port.TargetPort.IntVal
		}

		portName := port.Name
		if portName == "" {
			portName = "port-" + strconv.Itoa(i)
		}

		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Port:       servicePort,
			TargetPort: port.TargetPort,
			Protocol:   protocol,
			Name:       portName,
		})
	}

	return service
}

func generateBackendConfig(challenge *kctfv1.Challenge) *backendv1.BackendConfig {
	config := &backendv1.BackendConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: backendv1.BackendConfigSpec{
			SecurityPolicy: &backendv1.SecurityPolicyConfig{
				Name: os.Getenv("SECURITY_POLICY"),
			},
		},
	}
	return config
}

func findHTTPSPort(challenge *kctfv1.Challenge) *kctfv1.PortSpec {
	for _, port := range challenge.Spec.Network.Ports {
		// non-HTTPS is handled by generateLoadBalancerService
		if port.Protocol != "HTTPS" {
			continue
		}
		return &port
	}
	return nil
}

func generateManagedCertificate(challenge *kctfv1.Challenge, domains []string) *gkenetv1.ManagedCertificate {
	cert := &gkenetv1.ManagedCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
			Labels:    map[string]string{"app": challenge.Name},
		},
		Spec: gkenetv1.ManagedCertificateSpec{
			Domains: domains,
		},
		Status: gkenetv1.ManagedCertificateStatus{
			DomainStatus: []gkenetv1.DomainStatus{},
		},
	}
	return cert
}

func generateIngress(domainName string, challenge *kctfv1.Challenge, port *kctfv1.PortSpec) *netv1beta1.Ingress {
	// Ingress object
	ingress := &netv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        challenge.Name,
			Namespace:   challenge.Namespace,
			Labels:      map[string]string{"app": challenge.Name},
			Annotations: map[string]string{},
		},
		Spec: netv1beta1.IngressSpec{
			TLS: []netv1beta1.IngressTLS{{
				SecretName: "tls-cert",
			}},
			Rules: []netv1beta1.IngressRule{{
				Host: challenge.Name + "-web." + domainName,
			}},
		},
	}

	servicePort := port.Port
	if servicePort == 0 {
		servicePort = port.TargetPort.IntVal
	}

	ingress.Spec.Backend = &netv1beta1.IngressBackend{
		ServiceName: challenge.Name,
		ServicePort: intstr.FromInt(int(servicePort)),
	}

	if port.Domains != nil {
		ingress.Annotations["networking.gke.io/managed-certificates"] = challenge.Name
	}

	return ingress
}

func generateLoadBalancerService(domainName string, challenge *kctfv1.Challenge) *corev1.Service {
	// Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name + "-lb-service",
			Namespace: challenge.Namespace,
			Labels:    map[string]string{"app": challenge.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector:                 map[string]string{"app": challenge.Name},
			Type:                     "LoadBalancer",
			LoadBalancerSourceRanges: strings.Split(os.Getenv("ALLOWED_IPS"), ","),
		},
	}

	for i, port := range challenge.Spec.Network.Ports {
		// HTTPS is handled by generateIngress
		if port.Protocol == "HTTPS" {
			continue
		}

		servicePortNumber := port.Port
		if servicePortNumber == 0 {
			servicePortNumber = port.TargetPort.IntVal
		}

		servicePort := corev1.ServicePort{
			Port:       servicePortNumber,
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

	service.ObjectMeta.Annotations =
		map[string]string{"external-dns.alpha.kubernetes.io/hostname": challenge.Name + "." + domainName}

	return service
}
