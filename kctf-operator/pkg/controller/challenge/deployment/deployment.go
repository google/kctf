package deployment

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Deployment without Healthcheck
func deployment(challenge *kctfv1alpha1.Challenge, scheme *runtime.Scheme) *appsv1.Deployment {
	ls := labelsForChallenge(challenge.Name)
	var replicas int32 = 1
	var readOnlyRootFilesystem = true

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      challenge.Name,
			Namespace: challenge.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					/*Annotations: map[string]string{
						"container.apparmor.security.beta.kubernetes.io/challenge": "localhost/ctf-profile",
					},*/
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: challenge.Spec.ImageTemplate,
						Name:  "challenge",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{
									"SYS_ADMIN",
								},
							},
							ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
						},
						Command: []string{},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": *resource.NewMilliQuantity(900, resource.DecimalSI),
							},
							Requests: corev1.ResourceList{
								"cpu": *resource.NewMilliQuantity(450, resource.DecimalSI),
							},
						},
						// Uncomment when start testing with real challenges
						/*VolumeMounts: []corev1.VolumeMount{{
							Name:      "pow",
							ReadOnly:  true,
							MountPath: "/kctf/pow",
						},
							{
								Name:      "pow-bypass-pub",
								ReadOnly:  true,
								MountPath: "/kctf/pow-bypass",
							}},*/
					}},
					// Uncomment when start testing with real challenges
					/*Volumes: []corev1.Volume{{
						Name: "pow",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pow",
								},
							},
						},
					},
						{
							Name: "pow-bypass-pub",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "pow-bypass-pub",
								},
							},
						}},*/
				},
			},
		},
	}

	// Set container ports based on the ports that were passed
	dep.Spec.Template.Spec.Containers[0].Ports = ContainerPorts(challenge)
	// Merges with Pod Template
	if challenge.Spec.PodTemplate != nil {
		MergeWithPodTemplate(challenge, dep)
	}
	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, scheme)
	return dep
}
