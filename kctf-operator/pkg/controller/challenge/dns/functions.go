package dns

import (
	"context"

	netgkev1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1"
	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	utils "github.com/google/kctf/pkg/controller/challenge/utils"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func isEqual(certificateFound *netgkev1.ManagedCertificate,
	certificate *netgkev1.ManagedCertificate) bool {
	return certificateFound.Spec.Domains[0] == certificate.Spec.Domains[0]
}

func create(domainName string, challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates autoscaling if it doesn't exist yet
	certificate := generate(domainName, challenge)
	log.Info("Creating a Certificate", "Certificate name: ",
		certificate.Name, " with namespace ", certificate.Namespace)

	// We don't set a reference since we don't want this object to be garbage collected
	// Creating a certificate takes a long time, so keep it alive.
	// controllerutil.SetControllerReference(challenge, certificate, scheme)

	// Creates autoscaling
	err := client.Create(ctx, certificate)

	if err != nil {
		log.Error(err, "Failed to create Certificate", "Certificate name: ",
			certificate.Name, " with namespace ", certificate.Namespace)
		return false, err
	}

	return true, nil
}

func Update(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Creates certificate object
	existingCertificate := &netgkev1.ManagedCertificate{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, existingCertificate)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	certificateExists := err == nil

	existingIngress := &netv1beta1.Ingress{}
	err = client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, existingIngress)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}
	ingressExists := err == nil

	// We get the configmap that contains the domain name and get it
	domainName := utils.GetDomainName(challenge, client, log, ctx)

	log.Info("Certificate update status",
		"challengeName", challenge.Name,
		"certificateExists", certificateExists,
		"ingressExists", ingressExists,
		"domainName", domainName,
		"challenge.Spec.Network.Dns", challenge.Spec.Network.Dns)

	if !ingressExists || !challenge.Spec.Network.Dns || domainName == "" {
		// No certificate required.
		// Note that we don't delete the certificate here since creation takes a long time so we might want to reuse it in the future.
		return false, nil
	}

	// We checked that we want a certificate. Either create or update it.
	if !certificateExists {
		return create(domainName, challenge, client, scheme, log, ctx)
	}

	// Nothing to do if the certificates are the same
	newCertificate := generate(domainName, challenge)
	if isEqual(existingCertificate, newCertificate) {
		return false, nil
	}

	existingCertificate.Spec.Domains[0] = newCertificate.Spec.Domains[0]
	err = client.Update(ctx, existingCertificate)
	if err != nil {
		log.Error(err, "Failed to update certificate",
			"Certificate name: ", existingCertificate.Name,
			" with namespace ", existingCertificate.Namespace)
		return false, err
	}
	log.Info("Updated certificate successfully",
		"Certificate name: ", existingCertificate.Name,
		" with namespace ", existingCertificate.Namespace)
	return true, nil
}
