package pow

import (
	"strconv"

	kctfv1 "github.com/google/kctf/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Generates the configmap that contains how difficult should be the proof of work
func generate(challenge *kctfv1.Challenge) *corev1.ConfigMap {
	data := map[string]string{
		"pow.conf": strconv.Itoa(challenge.Spec.PowDifficultySeconds*1337) + "\n",
	}
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name + "-pow",
			Namespace: challenge.Namespace,
		},
		Data: data,
	}

	return configmap
}
