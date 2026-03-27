// Creates the service

package service

import (
	"context"
	"fmt"
	"reflect"

	gkenetv1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1"
	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	utils "github.com/google/kctf/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	backendv1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func annotationEqual(a, b map[string]string, key string) bool {
	return a[key] == b[key]
}

// isServiceEqual compares two services for equality.
// Used for both NodePort (internal) and LoadBalancer services — the annotation
// check is a no-op for NodePort since neither side sets annExternalDNSHostname.
func isServiceEqual(serviceFound *corev1.Service, serv *corev1.Service) bool {
	if !equalPorts(serviceFound.Spec.Ports, serv.Spec.Ports) {
		return false
	}
	if !reflect.DeepEqual(serviceFound.Spec.LoadBalancerSourceRanges, serv.Spec.LoadBalancerSourceRanges) {
		return false
	}
	return annotationEqual(serviceFound.Annotations, serv.Annotations, annExternalDNSHostname)
}

func isCertEqual(existingCert *gkenetv1.ManagedCertificate, newCert *gkenetv1.ManagedCertificate) bool {
	return reflect.DeepEqual(existingCert.Spec.Domains, newCert.Spec.Domains)
}

func isIngressEqual(ingressFound *netv1.Ingress, ingress *netv1.Ingress) bool {
	if !reflect.DeepEqual(ingressFound.Spec, ingress.Spec) {
		return false
	}
	return annotationEqual(ingressFound.Annotations, ingress.Annotations, annManagedCertificates)
}

// Check if the arrays of ports are the same
func equalPorts(found []corev1.ServicePort, wanted []corev1.ServicePort) bool {
	if len(found) != len(wanted) {
		return false
	}

	for i := range found {
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

func copyLoadBalancerSourceRanges(existingService *corev1.Service, newService *corev1.Service) {
	existingService.Spec.LoadBalancerSourceRanges = []string{}
	existingService.Spec.LoadBalancerSourceRanges = append(existingService.Spec.LoadBalancerSourceRanges, newService.Spec.LoadBalancerSourceRanges...)
}

func updateInternalService(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	newService := generateNodePortService(challenge)
	existingService := &corev1.Service{}

	err := client.Get(ctx, types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, existingService)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	serviceExists := err == nil

	if serviceExists {
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

func updateBackendConfig(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	existingConfig := &backendv1.BackendConfig{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name, Namespace: challenge.Namespace}, existingConfig)

	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	configExists := err == nil

	if configExists {
		// Currently, the config doesn't change. It always just points to the same security policy.
		// If we allow configuring more features, we will need to implement updating the existing config.
		return false, nil
	}

	newConfig := generateBackendConfig(challenge)

	controllerutil.SetControllerReference(challenge, newConfig, scheme)

	err = client.Create(ctx, newConfig)

	return true, err
}

func updateManagedCertificates(challenge *kctfv1.Challenge, c client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	desiredCerts := generateManagedCertificates(challenge)

	changed := false
	desiredNames := make(map[string]bool)

	for _, newCert := range desiredCerts {
		desiredNames[newCert.Name] = true

		existingCert := &gkenetv1.ManagedCertificate{}
		err := c.Get(ctx, types.NamespacedName{Name: newCert.Name, Namespace: newCert.Namespace}, existingCert)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		}

		if err == nil {
			if isCertEqual(existingCert, newCert) {
				continue
			}
			existingCert.Spec.Domains = newCert.Spec.Domains
			if err := c.Update(ctx, existingCert); err != nil {
				log.Error(err, "Failed to update managed certificate", "name", newCert.Name)
				return changed, err
			}
			log.Info("Updated managed certificate", "name", newCert.Name)
			changed = true
		} else {
			controllerutil.SetControllerReference(challenge, newCert, scheme)
			if err := c.Create(ctx, newCert); err != nil {
				return changed, err
			}
			log.Info("Created managed certificate", "name", newCert.Name)
			changed = true
		}
	}

	// Clean up stale certs
	existingCerts := &gkenetv1.ManagedCertificateList{}
	if err := c.List(ctx, existingCerts,
		client.InNamespace(challenge.Namespace),
		client.MatchingLabels{"app": challenge.Name}); err != nil {
		return changed, err
	}

	for i := range existingCerts.Items {
		cert := &existingCerts.Items[i]
		if desiredNames[cert.Name] {
			continue
		}
		if !metav1.IsControlledBy(cert, challenge) {
			continue
		}
		if err := c.Delete(ctx, cert); err != nil {
			log.Error(err, "Failed to delete stale managed certificate", "name", cert.Name)
			return changed, err
		}
		log.Info("Deleted stale managed certificate", "name", cert.Name)
		changed = true
	}

	return changed, nil
}

func updateIngresses(challenge *kctfv1.Challenge, c client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	domainName := utils.GetDomainName(challenge, c, log, ctx)

	var desiredIngresses []*netv1.Ingress
	if challenge.Spec.Network.Public {
		desiredIngresses = generateIngresses(domainName, challenge)
	}

	changed := false
	desiredNames := make(map[string]bool)

	for _, newIngress := range desiredIngresses {
		desiredNames[newIngress.Name] = true

		existingIngress := &netv1.Ingress{}
		err := c.Get(ctx, types.NamespacedName{Name: newIngress.Name, Namespace: newIngress.Namespace}, existingIngress)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		}

		if err == nil {
			if isIngressEqual(existingIngress, newIngress) {
				continue
			}
			existingIngress.Spec = newIngress.Spec
			existingIngress.ObjectMeta.Annotations = newIngress.ObjectMeta.Annotations
			if err := c.Update(ctx, existingIngress); err != nil {
				log.Error(err, "Failed to update ingress", " Name: ", newIngress.Name)
				return changed, err
			}
			log.Info("Updated ingress", " Name: ", newIngress.Name)
			changed = true
		} else {
			controllerutil.SetControllerReference(challenge, newIngress, scheme)
			if err := c.Create(ctx, newIngress); err != nil {
				return changed, err
			}
			log.Info("Created ingress", " Name: ", newIngress.Name)
			changed = true
		}
	}

	// Clean up stale ingresses
	existingIngresses := &netv1.IngressList{}
	if err := c.List(ctx, existingIngresses,
		client.InNamespace(challenge.Namespace),
		client.MatchingLabels{"app": challenge.Name}); err != nil {
		return changed, err
	}

	for i := range existingIngresses.Items {
		ing := &existingIngresses.Items[i]
		if desiredNames[ing.Name] {
			continue
		}
		if !metav1.IsControlledBy(ing, challenge) {
			continue
		}
		if err := c.Delete(ctx, ing); err != nil {
			log.Error(err, "Failed to delete stale ingress", " Name: ", ing.Name)
			return changed, err
		}
		log.Info("Deleted stale ingress", " Name: ", ing.Name)
		changed = true
	}

	return changed, nil
}

func updateLoadBalancerServices(challenge *kctfv1.Challenge, c client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	domainName := utils.GetDomainName(challenge, c, log, ctx)

	var desiredServices []*corev1.Service
	if challenge.Spec.Network.Public {
		desiredServices = generateLoadBalancerServices(domainName, challenge)
	}

	changed := false
	desiredNames := make(map[string]bool)

	// Create or update desired services
	for _, newService := range desiredServices {
		desiredNames[newService.Name] = true

		existingService := &corev1.Service{}
		err := c.Get(ctx, types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, existingService)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		}

		if err == nil {
			if isServiceEqual(existingService, newService) {
				continue
			}
			copyPorts(existingService, newService)
			existingService.ObjectMeta.Annotations = newService.ObjectMeta.Annotations
			copyLoadBalancerSourceRanges(existingService, newService)
			if err := c.Update(ctx, existingService); err != nil {
				log.Error(err, "Failed to update LB service", " Name: ", newService.Name)
				return changed, err
			}
			log.Info("Updated LB service", " Name: ", newService.Name)
			changed = true
		} else {
			controllerutil.SetControllerReference(challenge, newService, scheme)
			if err := c.Create(ctx, newService); err != nil {
				return changed, err
			}
			log.Info("Created LB service", " Name: ", newService.Name)
			changed = true
		}
	}

	// Clean up stale LB services
	existingServices := &corev1.ServiceList{}
	if err := c.List(ctx, existingServices,
		client.InNamespace(challenge.Namespace),
		client.MatchingLabels{"app": challenge.Name}); err != nil {
		return changed, err
	}

	for i := range existingServices.Items {
		svc := &existingServices.Items[i]
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}
		if desiredNames[svc.Name] {
			continue
		}
		if !metav1.IsControlledBy(svc, challenge) {
			continue
		}
		if err := c.Delete(ctx, svc); err != nil {
			log.Error(err, "Failed to delete stale LB service", " Name: ", svc.Name)
			return changed, err
		}
		log.Info("Deleted stale LB service", " Name: ", svc.Name)
		changed = true
	}

	return changed, nil
}

func checkPortsValid(challenge *kctfv1.Challenge) error {
	ports := make(map[int32]int32)
	httpsDomains := make(map[string]bool)

	for _, port := range challenge.Spec.Network.Ports {
		externalPort := port.Port
		targetPort := port.TargetPort.IntVal
		if externalPort == 0 {
			externalPort = targetPort
		}
		existingPort, portExists := ports[externalPort]
		if portExists && existingPort != targetPort {
			return fmt.Errorf("conflicting port mapping %v->%v and %v->%v", externalPort, existingPort, externalPort, targetPort)
		}
		ports[externalPort] = targetPort

		if port.Protocol == "HTTPS" {
			if httpsDomains[port.Domain] {
				if port.Domain == "" {
					return fmt.Errorf("only one HTTPS port allowed without an explicit domain")
				}
				return fmt.Errorf("duplicate domain %q for HTTPS ports", port.Domain)
			}
			httpsDomains[port.Domain] = true
		}
	}
	return nil
}

func Update(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	changed := false

	err := checkPortsValid(challenge)
	if err != nil {
		log.Error(err, "Invalid port configuration",
			" Name: ", challenge.Name,
			" with namespace ", challenge.Namespace)
		return false, err
	}

	internalServiceChanged, err := updateInternalService(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating internal service", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || internalServiceChanged

	loadBalancerServicesChanged, err := updateLoadBalancerServices(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating load balancer services", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || loadBalancerServicesChanged

	backendConfigChanged, err := updateBackendConfig(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating backend config for load balancer", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || backendConfigChanged

	managedCertsChanged, err := updateManagedCertificates(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating managed certificates", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || managedCertsChanged

	ingressesChanged, err := updateIngresses(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating ingresses", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || ingressesChanged

	return changed, nil
}
