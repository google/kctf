package dns

import (
	"context"

	netgkev1beta1 "github.com/GoogleCloudPlatform/gke-managed-certs/pkg/apis/networking.gke.io/v1beta1"
	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isEqual(certificateFound *netgkev1beta1.ManagedCertificate,
	certificate *netgkev1beta1.ManagedCertificate) bool {
	return certificateFound.Spec.Domains[0] == certificate.Spec.Domains[0]
}

func create(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates autoscaling if it doesn't exist yet
	certificate := generate(challenge)
	log.Info("Creating a Certificate", "Certificate name: ",
		certificate.Name, " with namespace ", certificate.Namespace)

	// Creates owner references
	controllerutil.SetControllerReference(challenge, certificate, scheme)

	// Creates autoscaling
	err := client.Create(ctx, certificate)

	if err != nil {
		log.Error(err, "Failed to create Certificate", "Certificate name: ",
			certificate.Name, " with namespace ", certificate.Namespace)
		return false, err
	}

	return true, nil
}

func delete(certificateFound *netgkev1beta1.ManagedCertificate, client client.Client,
	scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	log.Info("Deleting Certificate", "Certificate name: ",
		certificateFound.Name, " with namespace ", certificateFound.Namespace)

	err := client.Delete(ctx, certificateFound)
	if err != nil {
		log.Error(err, "Failed to delete Certificate", "Certificate name: ",
			certificateFound.Name, " with namespace ", certificateFound.Namespace)
		return false, err
	}

	return true, nil
}

func Update(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Creates certificate object
	certificateFound := &netgkev1beta1.ManagedCertificate{}
	ingressFound := &netv1beta1.Ingress{}
	err := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, certificateFound)
	err_ingress := client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, ingressFound)

	// First we check if there's any ingress (web challenge)
	if err_ingress == nil {
		// Then we check dns and domain name
		if challenge.Spec.Network.Dns == false {
			log.Info("Can't create certificate for web challenge, since DNS is disabled.")
		}

		if challenge.Spec.Network.DomainName == "" {
			log.Info("Can't create certificate for web challenge, since DomainName is empty.")
		}

		// If there's no certificate, we create one
		if challenge.Spec.Network.Dns == true && challenge.Spec.Network.DomainName != "" {
			if errors.IsNotFound(err) {
				return create(challenge, client, scheme, log, ctx)
			}

			// If there is, we update it if necessary
			if err == nil {
				if certificate := generate(challenge); !isEqual(certificateFound, certificate) {
					certificateFound.Spec.Domains[0] = certificate.Spec.Domains[0]
					err = client.Update(ctx, certificateFound)
					if err != nil {
						log.Error(err, "Failed to update certificate", "Certificate name: ",
							certificateFound.Name, " with namespace ", certificateFound.Namespace)
						return false, err
					}
					log.Info("Updated certificate successfully", "Certificate name: ",
						certificateFound.Name, " with namespace ", certificateFound.Namespace)
					return true, nil
				}
			}
		}
	}

	// If there's no ingress or dns/domainName is disabled/empty and there's a certificate, we delete it
	if err == nil && (errors.IsNotFound(err_ingress) ||
		challenge.Spec.Network.Dns == false || challenge.Spec.Network.DomainName == "") {
		return delete(certificateFound, client, scheme, log, ctx)
	}

	return false, nil
}
