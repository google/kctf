package pow

import (
	"strconv"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Generates the configmap that contains the how difficult should be the proof of work
func generate(challenge *kctfv1alpha1.Challenge) *corev1.ConfigMap {
	data := map[string]string{
		// TODO: Should it be like this? This is how it is in the configmap created
		"pow.conf": strconv.Itoa(challenge.Spec.PowDifficultySeconds) + "\n",
	}
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pow",
			Namespace: challenge.Namespace,
		},
		Data: data,
	}

	return configmap
}
