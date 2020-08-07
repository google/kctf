package initializer

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewDaemonSetGcsFuse() runtime.Object {
	privileged := true
	mountPropagation := corev1.MountPropagationBidirectional
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctf-daemon-gcsfuse",
			Namespace: "kube-system",
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
						Image: "ubuntu:19.10",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
						},
						Command: []string{"sh", "-c",
							"apt-get update && apt-get install -y wget fuse && wget -q https://github.com/GoogleCloudPlatform/gcsfuse/releases/download/v0.29.0/gcsfuse_0.29.0_amd64.deb && dpkg -i gcsfuse_0.29.0_amd64.deb && mkdir -p /mnt/disks/gcs && ((test -f /config/gcs_bucket && gcsfuse --foreground --debug_fuse --debug_gcs --stat-cache-ttl 0 -type-cache-ttl 0 -o allow_other -o nonempty --file-mode 0777 --dir-mode 0777 --uid 1000 --gid 1000 $(cat /config/gcs_bucket) /mnt/disks/gcs))"},
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
							PreStop: &corev1.Handler{
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
