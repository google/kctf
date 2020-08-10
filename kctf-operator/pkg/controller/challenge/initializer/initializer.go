package initializer

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func IsBeingInitialized(challenge *kctfv1alpha1.Challenge) bool {
	// TODO: can't use just nil, create a way to check this
	//isObjectMarkedAsInitialized := challenge.ObjectMeta.CreationTimestamp == nil
	return false
}

func InitializeChallenge(challenge *kctfv1alpha1.Challenge, client client.Client, log logr.Logger,
	ctx context.Context) (reconcile.Result, error) {

	log.Info("Creating namespace for this challenge")

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: challenge.Name,
		},
	}

	err := client.Create(ctx, namespace)

	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func FinalizerCleanPersistentVolumeClaim() {
	// When namespaced is deleted, this function will be called
	// Other option to do this is use a finalizer directly in the challenge
	// instead of being in the namespace
	// TODO: Clean the Persistent Volume Claim when the namespace is erased
}
