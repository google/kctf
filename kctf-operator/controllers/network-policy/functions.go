package network

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func isEqual(existingPolicy *netv1.NetworkPolicy, newPolicy *netv1.NetworkPolicy) bool {
	return reflect.DeepEqual(existingPolicy.Spec, newPolicy.Spec)
}

func Update(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	requeue := false
	var err error

	for _, policy := range generatePolicies(challenge) {
		requeue, err = updatePolicy(ctx, policy, challenge, cl, scheme, log)
		if err != nil {
			return false, err
		}
	}

	return requeue, nil
}

func updatePolicy(ctx context.Context, policy netv1.NetworkPolicy, challenge *kctfv1.Challenge,
	cl client.Client, scheme *runtime.Scheme, log logr.Logger) (bool, error) {
	existingPolicy := &netv1.NetworkPolicy{}
	err := cl.Get(ctx, types.NamespacedName{Name: policy.ObjectMeta.Name,
		Namespace: policy.ObjectMeta.Namespace}, existingPolicy)

	// Just enters here if it's a new policy
	if err != nil && errors.IsNotFound(err) {
		// Create a new object
		controllerutil.SetControllerReference(challenge, &policy, scheme)
		err = cl.Create(ctx, &policy)
		if err != nil {
			log.Error(err, "Failed to create Policy", " Name: ",
				policy.ObjectMeta.Name, " with namespace ", policy.ObjectMeta.Namespace)
			return false, err
		}
		return true, nil
	} else if err != nil {
		log.Error(err, "Couldn't get the Policy", " Name: ",
			policy.ObjectMeta.Name, " with namespace ", policy.ObjectMeta.Namespace)
		return false, err
	}

	if !isEqual(existingPolicy, &policy) {
		existingPolicy.Spec = policy.Spec
		err = cl.Update(ctx, existingPolicy)
		if err != nil {
			log.Error(err, "Failed to update Policy", " Name: ",
				policy.ObjectMeta.Name, " with namespace ", policy.ObjectMeta.Namespace)
			return false, err
		}

		log.Info("Policy updated succesfully", " Name: ",
			policy.ObjectMeta.Name, " with namespace ", policy.ObjectMeta.Namespace)
		return true, nil
	}

	return false, nil
}
