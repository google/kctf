// Creates deployment

package challenge

import (
	"context"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ContainerPorts(challenge *kctfv1alpha1.Challenge) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{}

	for _, port := range challenge.Spec.Network.Ports {
		containerPort := corev1.ContainerPort{
			ContainerPort: port.TargetPort.IntVal,
		}
		ports = append(ports, containerPort)
	}

	return ports
}

// Deployment with Healthcheck
func (r *ReconcileChallenge) deploymentWithHealthcheck(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	//TODO
	dep := &appsv1.Deployment{}
	return dep
}

// Deployment without Healthcheck
// TODO: Connect the podTemplate passed and the deployment
func (r *ReconcileChallenge) deploymentWithoutHealthcheck(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
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
						// TODO: "command" is present in the original deployment file but it's empty, should we add set it as nil?
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

	// Set Challenge instance as the owner and controller
	controllerutil.SetControllerReference(challenge, dep, r.scheme)
	return dep
}

// labelsForChallenge returns the labels for selecting the resources
// belonging to the given challenge CR name.
func labelsForChallenge(name string) map[string]string {
	return map[string]string{"app": "challenge", "challenge_cr": name}
}

// deploymentForChallenge returns a challenge Deployment object
func (r *ReconcileChallenge) deploymentForChallenge(challenge *kctfv1alpha1.Challenge) *appsv1.Deployment {
	if challenge.Spec.Healthcheck.Enabled == true {
		return r.deploymentWithHealthcheck(challenge)
	} else {
		return r.deploymentWithoutHealthcheck(challenge)
	}
}

func (r *ReconcileChallenge) CreateDeployment(challenge *kctfv1alpha1.Challenge,
	ctx context.Context) (reconcile.Result, error) {
	dep := r.deploymentForChallenge(challenge)
	r.log.Info("Creating a new Deployment", "Deployment.Namespace",
		dep.Namespace, "Deployment.Name", dep.Name)
	err := r.client.Create(ctx, dep)

	if err != nil {
		r.log.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
			dep.Namespace, "Deployment.Name", dep.Name)
		return reconcile.Result{}, err
	}

	// Deployment created successfully - return and requeue
	return reconcile.Result{Requeue: true}, nil
}
