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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	backendv1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isServiceEqual(serviceFound *corev1.Service, serv *corev1.Service) bool {
	if !equalPorts(serviceFound.Spec.Ports, serv.Spec.Ports) {
		return false
	}
	return reflect.DeepEqual(serviceFound.Spec.LoadBalancerSourceRanges, serv.Spec.LoadBalancerSourceRanges)
}

func isCertEqual(existingCert *gkenetv1.ManagedCertificate, newCert *gkenetv1.ManagedCertificate) bool {
	return reflect.DeepEqual(existingCert.Spec.Domains, newCert.Spec.Domains)
}

func isIngressEqual(ingressFound *netv1.Ingress, ingress *netv1.Ingress) bool {
	return reflect.DeepEqual(ingressFound.Spec, ingress.Spec)
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

func updateManagedCertificate(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	existingCert := &gkenetv1.ManagedCertificate{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name, Namespace: challenge.Namespace}, existingCert)

	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	certExists := err == nil

	port := findHTTPSPort(challenge)
	if port == nil || port.Domains == nil {
		if certExists {
			err := client.Delete(ctx, existingCert)
			return true, err
		}
		return false, nil
	}

	newCert := generateManagedCertificate(challenge, port.Domains)

	if certExists {
		if isCertEqual(existingCert, newCert) {
			return false, nil
		}

		existingCert.Spec.Domains = newCert.Spec.Domains

		err := client.Update(ctx, existingCert)

		return true, err
	}

	controllerutil.SetControllerReference(challenge, newCert, scheme)

	err = client.Create(ctx, newCert)

	return true, err
}

func updateIngress(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	existingIngress := &netv1.Ingress{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name, Namespace: challenge.Namespace}, existingIngress)

	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	ingressExists := err == nil

	port := findHTTPSPort(challenge)
	// Only one https port is supported at the moment.
	// To support more, we will need a field to specify the domain name per ingress.

	if port == nil {
		if ingressExists {
			err := client.Delete(ctx, existingIngress)
			return true, err
		}
		return false, nil
	}

	domainName := utils.GetDomainName(challenge, client, log, ctx)
	newIngress := generateIngress(domainName, challenge, port)

	if ingressExists {
		if isIngressEqual(existingIngress, newIngress) {
			return false, nil
		}

		existingIngress.Spec.DefaultBackend = newIngress.Spec.DefaultBackend
		existingIngress.ObjectMeta.Annotations = newIngress.ObjectMeta.Annotations
		err := client.Update(ctx, existingIngress)

		return true, err
	}

	if newIngress.Spec.DefaultBackend == nil || challenge.Spec.Network.Public == false {
		return false, nil
	}

	controllerutil.SetControllerReference(challenge, newIngress, scheme)

	err = client.Create(ctx, newIngress)

	return true, err
}

func updateLoadBalancerService(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Service is created in challenge_controller and here we just ensure that everything is alright
	// Creates the service if it doesn't exist
	// Check existence of the service:
	existingService := &corev1.Service{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name + "-lb-service",
		Namespace: challenge.Namespace}, existingService)

	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	serviceExists := err == nil

	// Get the domainName
	domainName := utils.GetDomainName(challenge, client, log, ctx)
	newService := generateLoadBalancerService(domainName, challenge)

	if serviceExists {
		if len(newService.Spec.Ports) == 0 || challenge.Spec.Network.Public == false {
			err := client.Delete(ctx, existingService)
			return true, err
		}

		if isServiceEqual(existingService, newService) {
			return false, nil
		}

		copyPorts(existingService, newService)
		existingService.ObjectMeta.Annotations = newService.ObjectMeta.Annotations
		copyLoadBalancerSourceRanges(existingService, newService)

		err := client.Update(ctx, existingService)

		if err == nil {
			log.Info("Updated load balancer service", " Name: ", newService.Name, " with namespace ", newService.Namespace)
		} else {
			log.Error(err, "Failed to update load balancer service", " Name: ", newService.Name, " with namespace ", newService.Namespace)
		}

		return true, err
	}

	if len(newService.Spec.Ports) == 0 || challenge.Spec.Network.Public == false {
		return false, nil
	}

	controllerutil.SetControllerReference(challenge, newService, scheme)

	err = client.Create(ctx, newService)

	return true, err
}

func checkPortsValid(challenge *kctfv1.Challenge) error {
	seenHTTPSPort := false
	ports := make(map[int32]int32)
	for _, port := range challenge.Spec.Network.Ports {
		if port.Protocol == "HTTPS" {
			if seenHTTPSPort {
				return fmt.Errorf("only one https port supported")
			}
		}
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

	loadBalancerServiceChanged, err := updateLoadBalancerService(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating load balancer service", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || loadBalancerServiceChanged

	backendConfigChanged, err := updateBackendConfig(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating backend config for load balancer", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || backendConfigChanged

	managedCertificateChanged, err := updateManagedCertificate(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating ManagedCertificate", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || managedCertificateChanged

	ingressChanged, err := updateIngress(challenge, client, scheme, log, ctx)
	if err != nil {
		log.Error(err, "Error updating ingress", " Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}
	changed = changed || ingressChanged

	return changed, nil
}
