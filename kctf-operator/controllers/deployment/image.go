package deployment

import (
	kctfv1 "github.com/google/kctf/api/v1"
	utils "github.com/google/kctf/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
)

func updateImages(deploymentFound *appsv1.Deployment, challenge *kctfv1.Challenge) bool {
	// Check if the image was changed and change it if necessary
	change := false
	idxChallenge := utils.IndexOfContainer("challenge", deploymentFound.Spec.Template.Spec.Containers)
	idxHealthcheck := utils.IndexOfContainer("healthcheck", deploymentFound.Spec.Template.Spec.Containers)

	if deploymentFound.Spec.Template.Spec.Containers[idxChallenge].Image != challenge.Spec.Image {
		deploymentFound.Spec.Template.Spec.Containers[idxChallenge].Image = challenge.Spec.Image
		change = true
	}
	if challenge.Spec.Healthcheck.Enabled == true {
		if deploymentFound.Spec.Template.Spec.Containers[idxHealthcheck].Image != challenge.Spec.Healthcheck.Image {
			deploymentFound.Spec.Template.Spec.Containers[idxHealthcheck].Image = challenge.Spec.Healthcheck.Image
			change = true
		}
	}

	return change
}
