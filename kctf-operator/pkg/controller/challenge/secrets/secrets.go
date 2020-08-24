package secrets

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generate(secretName string, challenge *kctfv1alpha1.Challenge,
	cl client.Client, scheme *runtime.Scheme, log logr.Logger,
	ctx context.Context) (*corev1.Secret, error) {
	// We get the secret from kube-system
	secretKube := &corev1.Secret{}
	err := cl.Get(ctx, types.NamespacedName{Name: secretName,
		Namespace: "kube-system"}, secretKube)

	if err != nil {
		return secretKube, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: challenge.Namespace,
		},

		Data: secretKube.Data,
	}

	return secret, nil
}
