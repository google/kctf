package utils

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// We get the configmap that contains the domain name and returns it
func GetDomainName(challenge *kctfv1.Challenge, client client.Client,
	log logr.Logger, ctx context.Context) string {
	domainName := ""
	configmap := &corev1.ConfigMap{}

	err := client.Get(ctx, types.NamespacedName{Name: "external-dns",
		Namespace: "kctf-system"}, configmap)

	if err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Couldn't get the configmap of the domain name.")
	}

	if err == nil {
		domainName = configmap.Data["DOMAIN_NAME"]
	}

	return domainName
}

// Find index of the container with a specific name in a list of containers
func IndexOfContainer(name string, containers []corev1.Container) int {
	for i, container := range containers {
		if container.Name == name {
			return i
		}
	}
	return -1
}
