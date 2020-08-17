package volumes

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreatePersistentVolumeClaim(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// TODO: Calls creation of persistent volume claim and persistent volume
	return false, nil
}

func DeletePersistentVolumeClaim(persistentVolumeClaimFound *corev1.PersistentVolumeClaim,
	client client.Client, scheme *runtime.Scheme, log logr.Logger,
	ctx context.Context) (bool, error) {
	// TODO: Calls deletion of persistent volume claim and persistent volume
	return false, nil
}
