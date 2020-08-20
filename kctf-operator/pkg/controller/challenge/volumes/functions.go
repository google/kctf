package volumes

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Calls creation of persistent volume claim and persistent volume
func Create(challenge *kctfv1alpha1.Challenge, claim string,
	client client.Client, scheme *runtime.Scheme, log logr.Logger, ctx context.Context) (bool, error) {
	pvc := persistentVolumeClaim(claim, challenge)

	// We set the ownership
	controllerutil.SetControllerReference(challenge, pvc, scheme)

	// First we create the persistent volume claim
	err := client.Create(ctx, pvc)
	if err != nil {
		log.Error(err, "Failed to create persistentVolumeClaim: ", "Name: ", pvc.Name,
			"Namespace: ", pvc.Namespace)
		return false, err
	}

	pv := persistentVolume(pvc, challenge)
	// We set the ownership
	controllerutil.SetControllerReference(challenge, pv, scheme)

	err = client.Create(ctx, pv)

	if err != nil {
		log.Error(err, "Failed to create persistentVolume: ", "Name: ", pv.Name, "Namespace: ",
			pv.Namespace)
		return false, err
	}

	return true, nil
}

func Delete(persistentVolumeClaim *corev1.PersistentVolumeClaim,
	client client.Client, scheme *runtime.Scheme, log logr.Logger,
	ctx context.Context) (bool, error) {
	// Calls deletion of persistent volume claim
	err := client.Delete(ctx, persistentVolumeClaim)

	// Calls deletion of persistent volume claim
	if err != nil {
		log.Error(err, "Failed to delete persistentVolumeClaim: ", "Name: ", persistentVolumeClaim.Name,
			"Namespace: ", persistentVolumeClaim.Namespace)
		return false, err
	}

	persistentVolume := &corev1.PersistentVolume{}
	err = client.Get(ctx, types.NamespacedName{Name: persistentVolumeClaim.Name,
		Namespace: persistentVolumeClaim.Namespace}, persistentVolume)

	if err != nil {
		log.Error(err, "Failed to get persistentVolume: ", "Name: ", persistentVolumeClaim.Name,
			"Namespace: ", persistentVolumeClaim.Namespace)
		return false, err
	}

	err = client.Delete(ctx, persistentVolume)

	if err != nil {
		log.Error(err, "Failed to delete persistentVolume: ", "Name: ", persistentVolumeClaim.Name,
			"Namespace: ", persistentVolumeClaim.Namespace)
		return false, err
	}

	return true, nil
}
