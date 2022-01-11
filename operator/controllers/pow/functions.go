package pow

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

func isEqual(configmapFound *corev1.ConfigMap, configmap *corev1.ConfigMap) bool {
	return reflect.DeepEqual(configmapFound.Data,
		configmap.Data)
}

// Create the configmaps
func create(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// creates pow if it doesn't exist yet
	configmap := generate(challenge)
	log.Info("Creating a ConfigMap for Proof of work", "ConfigMap name: ",
		configmap.Name, " with namespace ", configmap.Namespace)

	// Creates owner references
	controllerutil.SetControllerReference(challenge, configmap, scheme)

	// Creates configmap
	err := client.Create(ctx, configmap)

	if err != nil {
		log.Error(err, "Failed to create ConfigMap for Proof of work", "ConfigMap name: ",
			configmap.Name, " with namespace ", configmap.Namespace)
		return false, err
	}

	return true, nil
}

func Update(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	configmapFound := &corev1.ConfigMap{}
	err := cl.Get(ctx, types.NamespacedName{Name: challenge.Name + "-pow",
		Namespace: challenge.Namespace}, configmapFound)

	// Just enters here if it's a new configmap
	if err != nil && errors.IsNotFound(err) {
		// Create a new configmap
		return create(challenge, cl, scheme, log, ctx)

	} else if err != nil {
		log.Error(err, "Couldn't get the ConfigMap of Proof of work", "Configmap Name: ",
			challenge.Name+"-pow", " with namespace ", challenge.Namespace)
		return false, err
	}

	// Checks if the confimap is correctly set
	if configmap := generate(challenge); !isEqual(configmapFound, configmap) {
		configmapFound.Data = configmap.Data
		err = cl.Update(ctx, configmapFound)
		if err != nil {
			log.Error(err, "Failed to update ConfigMap for Proof of work", "ConfigMap Name: ",
				"pow", " with namespace ", challenge.Namespace)
			return false, err
		}
		log.Info("ConfigMap for Proof of Work updated succesfully", "Name: ",
			"pow", " with namespace ", challenge.Namespace)
		return true, nil
	}

	return false, nil
}
