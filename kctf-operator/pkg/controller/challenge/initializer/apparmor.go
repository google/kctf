package initializer

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewApparmorProfiles() runtime.Object {
	ctfProfile := `|-
    #include <tunables/global>
    profile ctf-profile flags=(attach_disconnected,mediate_deleted) {
      #include <abstractions/base>
      ptrace peer=@{profile_name},
      network,
      capability,
      file,
      mount,
      umount,
      pivot_root,
      deny @{PROC}/* w,  # deny write for all files directly in /proc (not in a subdir)
      # deny write to files not in /proc/<number>/** or /proc/sys/**
      deny @{PROC}/{[^1-9],[^1-9][^0-9],[^1-9s][^0-9y][^0-9s],[^1-9][^0-9][^0-9][^0-9]*}/** w,
      deny @{PROC}/sys/[^k]** w,  # deny /proc/sys except /proc/sys/k* (effectively /proc/sys/kernel)
      deny @{PROC}/sys/kernel/{?,??,[^s][^h][^m]**} w,  # deny everything except shm* in /proc/sys/kernel/
      deny @{PROC}/sysrq-trigger rwklx,
      deny @{PROC}/kcore rwklx,
      deny @{PROC}/mem rwklx,
      deny @{PROC}/kmem rwklx,
      deny /sys/[^f]*/** wklx,
      deny /sys/f[^s]*/** wklx,
      deny /sys/fs/[^c]*/** wklx,
      deny /sys/fs/c[^g]*/** wklx,
      deny /sys/fs/cg[^r]*/** wklx,
      deny /sys/firmware/** rwklx,
      deny /sys/kernel/security/** rwklx,
    }`
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apparmor-profiles",
			Namespace: "kube-system",
		},
		Data: map[string]string{"ctf-profile": ctfProfile},
	}
	return configmap
}
