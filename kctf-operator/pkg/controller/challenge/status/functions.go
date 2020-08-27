package status

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Update(requeue bool, err error, challenge *kctfv1alpha1.Challenge, cl client.Client,
	log logr.Logger, ctx context.Context) error {
	// First we update Status
	if err == nil && requeue == true {
		challenge.Status.Status = "updating"
	}

	if err == nil && requeue == false {
		challenge.Status.Status = "up-to-date"
	}

	if err != nil {
		challenge.Status.Status = "error"
	}

	// Then we update Health
	if challenge.Spec.Healthcheck.Enabled == false {
		challenge.Status.Health = "disabled"
		err = cl.Status().Update(ctx, challenge)
	} else {
		// check healthcheck
	}

	err_status := cl.Status().Update(ctx, challenge)

	if err_status != nil {
		log.Error(err, "Error updating status")
	}

	return err_status
}
