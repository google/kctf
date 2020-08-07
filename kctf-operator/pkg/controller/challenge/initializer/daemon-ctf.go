package initializer

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewDaemonSetCtf() runtime.Object {
	privileged := true
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-daemon",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "ctf-daemon"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "ctf-daemon"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": "ctf-daemon"},
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{{
						Key:    "node-role.kubernetes.io/master",
						Effect: corev1.TaintEffectNoSchedule,
					}},
					Containers: []corev1.Container{{
						Name:  "ctf-daemon",
						Image: "eu.gcr.io/google_containers/apparmor-loader",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
						},
						// TODO: Is this command correctly formated?
						Command: []string{"sh", "-c",
							"while true; do for f in /profiles/*; do echo \"loading $f\"; apparmor_parser -r $f; sleep 30; done; done"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "sys",
								MountPath: "/sys",
								ReadOnly:  true,
							},
							{
								Name:      "apparmor-includes",
								MountPath: "/etc/apparmor.d",
								ReadOnly:  true,
							},
							{
								Name:      "profiles",
								MountPath: "/profiles",
								ReadOnly:  true,
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "sys",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys",
								},
							},
						},
						{
							Name: "apparmor-includes",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/apparmor.d",
								},
							},
						},
						{
							Name: "profiles",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "apparmor-profiles",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return daemonSet
}
