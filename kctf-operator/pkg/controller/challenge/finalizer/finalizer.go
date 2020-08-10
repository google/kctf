// TODO; Finalizer to make the clean up after deletion
// TODO: should we create other controller to implement the clean up of the volumeclaims or leave it here?

package finalizer

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Checks if the object is marked to be deleted
func IsBeingFinalized(challenge *kctfv1alpha1.Challenge) bool {
	IsObjMarkedToBeDeleted := challenge.ObjectMeta.DeletionTimestamp != nil
	return IsObjMarkedToBeDeleted
}

func FinalizeChallenge(challenge *kctfv1alpha1.Challenge) (reconcile.Result, error) {
	// TODO
	return reconcile.Result{}, nil
}
