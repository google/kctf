package secrets

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isEqual(secretFound *corev1.Secret, secret *corev1.Secret) bool {
	return reflect.DeepEqual(secretFound.Data, secret.Data)
}

// Create the secrets
func create(secretName string, challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {

	secret, err := generate(secretName, challenge,
		client, scheme, log, ctx)

	if err != nil {
		log.Error(err, "Couldn't get the Secret from kctf-system", "Secret Name: ",
			secretName, " with namespace ", challenge.Namespace)
		return false, err
	}

	log.Info("Creating Secret", "Secret ", secret.Name,
		" with namespace ", challenge.Namespace)

	// Creates owner references
	controllerutil.SetControllerReference(challenge, secret, scheme)

	err = client.Create(ctx, secret)
	if err != nil {
		log.Error(err, "Failed to create Secret", "Secret name: ",
			secret.Name, " with namespace ", challenge.Namespace)
		return false, err
	}

	return true, nil
}

func Update(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	secrets := []string{"pow-bypass", "pow-bypass-pub", "tls-cert"}
	requeue := false
	var err error

	for _, secret := range secrets {
		// Creates object
		requeue, err = updateSecret(secret, challenge, cl, scheme, log, ctx)
		if err != nil {
			return false, err
		}
	}

	return requeue, nil
}

func updateSecret(secretName string, challenge *kctfv1.Challenge,
	cl client.Client, scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	secretFound := &corev1.Secret{}
	err := cl.Get(ctx, types.NamespacedName{Name: secretName,
		Namespace: challenge.Namespace}, secretFound)

	// Just enters here if it's a new secret
	if err != nil && errors.IsNotFound(err) {
		// Create a new secret
		return create(secretName, challenge, cl, scheme, log, ctx)

	} else if err != nil {
		log.Error(err, "Couldn't get the Secret", "Secret Name: ",
			secretName, " with namespace ", challenge.Namespace)
		return false, err
	}

	// Checks if the confimap and the secrets are correctly set
	secret, err := generate(secretName, challenge,
		cl, scheme, log, ctx)

	if err != nil {
		log.Error(err, "Couldn't get the Secret from kctf-system", "Secret Name: ",
			secretName, " with namespace ", challenge.Namespace)
		return false, err
	}

	if !isEqual(secretFound, secret) {
		secretFound.Data = secret.Data
		err = cl.Update(ctx, secretFound)
		if err != nil {
			log.Error(err, "Failed to update Secret", "Secret Name: ",
				secretName, " with namespace ", challenge.Namespace)
			return false, err
		}

		log.Info("Secret updated succesfully", "Name: ",
			secretName, " with namespace ", challenge.Namespace)
		return true, nil
	} else {
		log.Info("Secrets are the same", "name", secretName, "namespace", challenge.Namespace)
	}

	return false, nil
}
