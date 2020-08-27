package deployment

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

func updateImages(deploymentFound *appsv1.Deployment, challenge *kctfv1alpha1.Challenge) bool {
	// Check if the image was changed and change it if necessary
	change := false
	idx_challenge := find_idx("challenge", deploymentFound.Spec.Template.Spec.Containers)
	idx_healthcheck := find_idx("healthcheck", deploymentFound.Spec.Template.Spec.Containers)

	if deploymentFound.Spec.Template.Spec.Containers[idx_challenge].Image != challenge.Spec.Image {
		deploymentFound.Spec.Template.Spec.Containers[idx_challenge].Image = challenge.Spec.Image
		change = true
	}
	if challenge.Spec.Healthcheck.Enabled == true {
		if deploymentFound.Spec.Template.Spec.Containers[idx_challenge].Image != challenge.Spec.Image {
			deploymentFound.Spec.Template.Spec.Containers[idx_healthcheck].Image = challenge.Spec.Healthcheck.Image
			change = true
		}
	}

	return change
}
