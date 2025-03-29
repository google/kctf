package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDaemonSetGcsFuse() client.Object {
	privileged := true
	mountPropagation := corev1.MountPropagationBidirectional
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-daemon-gcsfuse",
			Namespace: "kctf-system",
			Labels:    map[string]string{"k8s-app": "ctf-daemon-gcsfuse"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "ctf-daemon-gcsfuse"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": "ctf-daemon-gcsfuse"},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "gcsfuse-sa",
					Tolerations: []corev1.Toleration{{
						Key:    "node-role.kubernetes.io/master",
						Effect: corev1.TaintEffectNoSchedule,
					}},
					Containers: []corev1.Container{{
						Name:  "ctf-daemon",
						Image: DOCKER_GCSFUSE_IMAGE,
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:             "mnt-disks-gcs",
								MountPath:        "/mnt/disks/gcs",
								MountPropagation: &mountPropagation,
							},
							{
								Name:      "config",
								MountPath: "/config",
							},
						},
						Lifecycle: &corev1.Lifecycle{
							PreStop: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"sh", "-c", "fusermount -u /mnt/disks/gcs"},
								},
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "mnt-disks-gcs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disks/gcs",
								},
							},
						},
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "gcsfuse-config",
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
