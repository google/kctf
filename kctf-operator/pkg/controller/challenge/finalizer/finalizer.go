// Finalizer to make the clean up after deletion
// TODO: should we create other controller to implement the clean up of the volumeclaims or leave it here?

package finalizer

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	utils "github.com/google/kctf/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ChallengeFinalizerName = "finalizer.kctf.v1alpha1"

// Checks if the object is marked to be deleted
func IsBeingFinalized(challenge *kctfv1alpha1.Challenge) bool {
	IsObjMarkedToBeDeleted := challenge.ObjectMeta.DeletionTimestamp != nil
	return IsObjMarkedToBeDeleted
}

func CallChallengeFinalizers(client client.Client, ctx context.Context, reqLogger logr.Logger,
	challenge *kctfv1alpha1.Challenge) (reconcile.Result, error) {
	if utils.Contains(challenge.GetFinalizers(), ChallengeFinalizerName) {
		// Run finalization logic for memcachedFinalizer. If the
		// finalization logic fails, don't remove the finalizer so
		// that we can retry during the next reconciliation.
		if err := FinalizeChallenge(client, ctx, reqLogger, challenge); err != nil {
			return reconcile.Result{}, err
		}

		// Remove ChallengeFinalizerName. Once all finalizers have been
		// removed, the object will be deleted.
		controllerutil.RemoveFinalizer(challenge, ChallengeFinalizerName)
		err := client.Update(ctx, challenge)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func FinalizeChallenge(client client.Client, ctx context.Context, reqLogger logr.Logger,
	challenge *kctfv1alpha1.Challenge) error {
	// Cleanup steps that the operator
	// needs to do before the CR can be deleted.
	// Deletes the namespace and cleans Persistent Volume Claim
	// TODO: Clean Persistent Volume Claim

	namespace := &corev1.Namespace{}

	client.Get(ctx, types.NamespacedName{Name: challenge.Name}, namespace)

	err := client.Delete(ctx, namespace)

	if err != nil {
		reqLogger.Info("Failed to delete challenge correctly")
		return err
	}

	reqLogger.Info("Successfully finalized challenge")
	return nil
}

func AddFinalizer(client client.Client, ctx context.Context, reqLogger logr.Logger,
	challenge *kctfv1alpha1.Challenge) error {
	reqLogger.Info("Adding Finalizer for the Challenge")
	controllerutil.AddFinalizer(challenge, ChallengeFinalizerName)

	// Update CR
	err := client.Update(ctx, challenge)
	if err != nil {
		reqLogger.Error(err, "Failed to update Challenge with finalizer")
		return err
	}
	return nil
}
