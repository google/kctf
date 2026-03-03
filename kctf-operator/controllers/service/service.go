package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	gkenetv1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1"
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	backendv1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1"
)

const (
	annExternalDNSHostname = "external-dns.alpha.kubernetes.io/hostname"
	annManagedCertificates = "networking.gke.io/managed-certificates"
)

func portHost(challengeName string, port *kctfv1.PortSpec, domainName string, defaultSuffix string) string {
	if port.Domain != "" {
		return port.Domain + "." + domainName
	}
	return challengeName + defaultSuffix + "." + domainName
}

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
		protocol := corev1.ProtocolTCP
		switch port.Protocol {
		case corev1.ProtocolSCTP, corev1.ProtocolTCP, corev1.ProtocolUDP:
			protocol = port.Protocol
		}

		servicePort := port.Port
		if servicePort == 0 {
			servicePort = port.TargetPort.IntVal
		}
		if portsSeen[servicePort] {
			continue
		}
		portsSeen[servicePort] = true

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
		Spec: backendv1.BackendConfigSpec{},
	}
	if os.Getenv("SECURITY_POLICY") != "DISABLED" {
		config.Spec.SecurityPolicy = &backendv1.SecurityPolicyConfig{
			Name: os.Getenv("SECURITY_POLICY"),
		}
	}
	return config
}

func ingressName(challengeName string, domain string) string {
	if domain == "" {
		return challengeName
	}
	return challengeName + "-ingress-" + domain
}

func certName(challengeName string, domain string) string {
	if domain == "" {
		return challengeName
	}
	return challengeName + "-cert-" + domain
}

func generateIngresses(domainName string, challenge *kctfv1.Challenge) []*netv1.Ingress {
	var ingresses []*netv1.Ingress

	for _, port := range challenge.Spec.Network.Ports {
		if port.Protocol != "HTTPS" {
			continue
		}

		servicePort := port.Port
		if servicePort == 0 {
			servicePort = port.TargetPort.IntVal
		}

		ingress := &netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        ingressName(challenge.Name, port.Domain),
				Namespace:   challenge.Namespace,
				Labels:      map[string]string{"app": challenge.Name},
				Annotations: map[string]string{},
			},
			Spec: netv1.IngressSpec{
				TLS: []netv1.IngressTLS{{
					SecretName: "tls-cert",
				}},
				Rules: []netv1.IngressRule{{
					Host: portHost(challenge.Name, &port, domainName, "-web"),
				}},
				DefaultBackend: &netv1.IngressBackend{
					Service: &netv1.IngressServiceBackend{
						Name: challenge.Name,
						Port: netv1.ServiceBackendPort{
							Number: int32(servicePort),
						},
					},
				},
			},
		}

		if port.Domains != nil {
			ingress.Annotations[annManagedCertificates] = certName(challenge.Name, port.Domain)
		}

		ingresses = append(ingresses, ingress)
	}

	return ingresses
}

func generateManagedCertificates(challenge *kctfv1.Challenge) []*gkenetv1.ManagedCertificate {
	var certs []*gkenetv1.ManagedCertificate

	for _, port := range challenge.Spec.Network.Ports {
		if port.Protocol != "HTTPS" || port.Domains == nil {
			continue
		}

		certs = append(certs, &gkenetv1.ManagedCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certName(challenge.Name, port.Domain),
				Namespace: challenge.Namespace,
				Labels:    map[string]string{"app": challenge.Name},
			},
			Spec: gkenetv1.ManagedCertificateSpec{
				Domains: port.Domains,
			},
			Status: gkenetv1.ManagedCertificateStatus{
				DomainStatus: []gkenetv1.DomainStatus{},
			},
		})
	}

	return certs
}

func lbServiceName(challengeName string, domain string) string {
	if domain == "" {
		return challengeName + "-lb-service"
	}
	return challengeName + "-lb-" + domain
}

func generateLoadBalancerServices(domainName string, challenge *kctfv1.Challenge) []*corev1.Service {
	type lbGroup struct {
		hostname string
		ports    []corev1.ServicePort
	}

	groups := make(map[string]*lbGroup)
	var groupOrder []string

	for i, port := range challenge.Spec.Network.Ports {
		if port.Protocol == "HTTPS" {
			continue
		}

		key := port.Domain
		group, ok := groups[key]
		if !ok {
			group = &lbGroup{
				hostname: portHost(challenge.Name, &port, domainName, ""),
			}
			groups[key] = group
			groupOrder = append(groupOrder, key)
		}

		servicePortNumber := port.Port
		if servicePortNumber == 0 {
			servicePortNumber = port.TargetPort.IntVal
		}

		portName := port.Name
		if portName == "" {
			portName = "port-" + strconv.Itoa(i)
		}

		group.ports = append(group.ports, corev1.ServicePort{
			Port:       servicePortNumber,
			TargetPort: port.TargetPort,
			Protocol:   port.Protocol,
			Name:       portName,
		})
	}

	var services []*corev1.Service
	for _, key := range groupOrder {
		group := groups[key]
		services = append(services, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      lbServiceName(challenge.Name, key),
				Namespace: challenge.Namespace,
				Labels:    map[string]string{"app": challenge.Name},
				Annotations: map[string]string{
					annExternalDNSHostname: group.hostname,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector:                 map[string]string{"app": challenge.Name},
				Type:                     "LoadBalancer",
				LoadBalancerSourceRanges: strings.Split(os.Getenv("ALLOWED_IPS"), ","),
				Ports:                    group.ports,
			},
		})
	}

	return services
}
