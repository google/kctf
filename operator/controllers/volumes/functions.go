package volumes

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Function that maps names of the persistent volume claims in the list to their index
func mapNameIdx(persistentVolumeClaimsFound *corev1.PersistentVolumeClaimList) map[string]int {
	m := make(map[string]int)

	for idx, item := range persistentVolumeClaimsFound.Items {
		m[item.Name] = idx
	}

	return m
}

// Calls creation of persistent volume claim and persistent volume
func create(challenge *kctfv1.Challenge, claim string,
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

// Name delete was changed to avoid being the same name as the function delete used to delete an element in
// the map
// This function delete the persistentVolumeClaim and the persistentVolume associated
func deleteVolumes(persistentVolumeClaim *corev1.PersistentVolumeClaim,
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

// Function that updates the persistent volume claim list and the persistent volumes
func Update(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// Check if all persistent volume claims are correctly set and update them if necessary
	// We get all persistentVolumeClaims in the same namespace as the challenge
	persistentVolumeClaimsFound := &corev1.PersistentVolumeClaimList{}
	change := false

	// List all persistent volume claims in the namespace of the challenge
	var listOption client.ListOption
	listOption = &client.ListOptions{
		Namespace:     challenge.Namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": challenge.Name}),
	}

	err := cl.List(ctx, persistentVolumeClaimsFound, listOption)
	if err != nil {
		log.Error(err, "Failed to list persistent volume claims", "Challenge Name: ",
			challenge.Name, " with namespace ", challenge.Namespace)
		return false, err
	}

	// First we create a map with the names of the persistent volume claims that already exist
	namesFound := mapNameIdx(persistentVolumeClaimsFound)

	// For comparing two persistentVolumeClaims, we will use DeepEqual
	if challenge.Spec.PersistentVolumeClaims != nil {
		for _, claim := range challenge.Spec.PersistentVolumeClaims {
			_, present := namesFound[claim]
			if present == true {
				delete(namesFound, claim)
			} else {
				// Creates the object
				change, err = create(challenge, claim,
					cl, scheme, log, ctx)
				if err != nil {
					return false, err
				}
				log.Info("PersistentVolumeClaim and PersistentVolume created successfully",
					"Name: ", claim, "Namespace:", challenge.Namespace)
			}
		}
	}

	// Then we delete the persistent volume claims that remained
	for name, idx := range namesFound {
		change, err = deleteVolumes(&persistentVolumeClaimsFound.Items[idx],
			cl, scheme, log, ctx)
		if err != nil {
			return false, err
		}
		log.Info("PersistentVolumeClaim and PersistentVolume deleted successfully",
			"Name: ", name, "Namespace:", challenge.Namespace)
	}

	return change, err
}
