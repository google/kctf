package pow

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Create the configmaps
// TODO: Do we create the secrets here?

func Create(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates pow if it doesn't exist yet
	configmap := Generate(challenge)
	log.Info("Creating a ConfigMap for Proof of work", "ConfigMap name: ",
		configmap.Name, " with namespace ", configmap.Namespace)

	// Creates owner references
	err := controllerutil.SetControllerReference(challenge, configmap, scheme)

	// Creates configmap
	err = client.Create(ctx, configmap)

	if err != nil {
		log.Error(err, "Failed to create ConfigMap for Proof of work", "ConfigMap name: ",
			configmap.Name, " with namespace ", configmap.Namespace)
		return false, err
	}

	return true, nil
}
