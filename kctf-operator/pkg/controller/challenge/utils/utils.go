package utils

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// We get the configmap that contains the domain name and returns it
func GetDomainName(challenge *kctfv1alpha1.Challenge, client client.Client,
	log logr.Logger, ctx context.Context) string {
	domainName := ""
	configmap := &corev1.ConfigMap{}

	err := client.Get(ctx, types.NamespacedName{Name: "external-dns",
		Namespace: "kube-system"}, configmap)

	if err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Couldn't get the configmap of the domain name.")
	}

	if err == nil {
		domainName = configmap.Data["DOMAIN_NAME"]
	}

	return domainName
}
