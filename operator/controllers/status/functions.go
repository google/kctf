package status

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	utils "github.com/google/kctf/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Update(requeue bool, err error, challenge *kctfv1.Challenge, cl client.Client,
	log logr.Logger, ctx context.Context) error {

	pods := &corev1.PodList{}
	var listOption client.ListOption
	listOption = &client.ListOptions{
		Namespace:     challenge.Namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": challenge.Name}),
	}

	err_list := cl.List(ctx, pods, listOption)

	if err_list == nil {
		// First we find the right pod
		for _, pod := range pods.Items {
			idx_challenge := utils.IndexOfContainer("challenge", pod.Spec.Containers)
			idx_healthcheck := utils.IndexOfContainer("healthcheck", pod.Spec.Containers)

			// This variable tells if the container is right one considering the healthcheck only
			right_healthcheck := !challenge.Spec.Healthcheck.Enabled

			// We prevent to get an error if the pod is being terminated
			if len(pod.Spec.Containers) != 0 {
				if right_healthcheck == false && idx_healthcheck != -1 {
					if pod.Spec.Containers[idx_healthcheck].Image != "healthcheck" {
						right_healthcheck = true
					}
				}
				// We take the right pod (it's possible that, if the challenge is not healthy,
				// that we have multiple pods)
				if idx_challenge > -1 && pod.Spec.Containers[idx_challenge].Image != "challenge" && right_healthcheck {
					// We update the status
					challenge.Status.Status = pod.Status.Phase

					// Then we update Health
					if challenge.Spec.Healthcheck.Enabled == false || idx_challenge >= len(pod.Status.ContainerStatuses) {
						challenge.Status.Health = "disabled"
					} else {
						// We check if the challenge is ready to know if it's healthy
						if pod.Status.ContainerStatuses[idx_challenge].Ready == false {
							challenge.Status.Health = "unhealthy"
						} else {
							challenge.Status.Health = "healthy"
						}
					}
				}
			}
		}
	} else {
		log.Error(err_list, "Failed to get pods")
	}

	err_status := cl.Status().Update(ctx, challenge)

	if err_status != nil {
		log.Error(err_status, "Error updating status")
	}

	return err_status
}
