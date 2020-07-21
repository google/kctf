// +kubebuilder:validation:Required
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Functions to return the default values
func PersistentVolumeClaimDefault() corev1.PersistentVolumeClaim {
	var persistentVolumeClaimDefault = corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": *resource.NewQuantity(1*1024*1024*1024*1024, resource.BinarySI),
				},
			},
		},
	}
	return persistentVolumeClaimDefault
}

func PodTemplateDefault() corev1.PodTemplate {
	var podTemplateDefault = corev1.PodTemplate{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				InitContainers: nil,
				Containers:     nil,
			},
		},
	}
	return podTemplateDefault
}

func HealthcheckDefault() HealthcheckSpec {
	var healthcheckDefault = HealthcheckSpec{
		Enabled:   false,
		Container: "healthcheck",
	}
	return healthcheckDefault
}

func NetworkDefault() NetworkSpec {
	var networkDefault = NetworkSpec{
		Public: true,
		Dns:    false,
		Ports:  PortsDefault(),
	}
	return networkDefault
}

func PortsDefault() []PortSpec {
	var portsDefault = []PortSpec{
		PortSpec{
			Port:       1,
			TargetPort: 1337,
			Protocol:   "TCP",
		},
	}
	return portsDefault
}

type PortSpec struct {

	// The default value is 1 if empty and protocol not being HTTP
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65336
	Port int32 `json:"port,omitempty"`

	// +kubebuider:default:=1337
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65336
	TargetPort int32 `json:"targetPort,omitempty"`

	// +kubebuilder:default:="TCP"
	Protocol string `json:"protocol,omitempty"`
}

type NetworkSpec struct {

	// +kubebuilder:default:=true
	Public bool `json:"public,omitempty"`

	// +kubebuilder:default:=false
	Dns bool `json:"dns,omitempty"`

	// By default, one port is set with default values
	// +kubebuilder:default:PortsDefault()
	Ports []PortSpec `json:"ports,omitempty"`
}

type HealthcheckSpec struct {

	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:default:="healthcheck"
	Container string `json:"container,omitempty"`
}

type AutoscalingSpec struct {

	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:default:=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:default:=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// If empty, this feature won't be used
	// +kubebuilder:validation:Optional
	TargetCPUUtilizationPercentage int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}

// TODO: create functions that return default values for this
type DeploymentSpec struct {

	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// TODO default
	// Default value: 1 container and 1 volume with the name of the challenge
	Template corev1.PodTemplate `json:"podTemplate,omitempty"` // TODO: https://github.com/kubernetes-sigs/controller-tools/issues/444
}

// ChallengeSpec defines the desired state of Challenge
type ChallengeSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// description
	// Not optional and image should be passed by user (by now)
	ImageTemplate string `json:"imageTemplate"`

	// +kubebuilder:default:=false
	Deployed bool `json:"deployed,omitempty"`

	// +kubebuilder:default:=0
	PowDifficultySeconds int32 `json:"powDifficultySeconds,omitempty"`

	// +kubebuilder:default:NetworkDefault()
	Network NetworkSpec `json:"network,omitempty"`

	// If empty, healthcheck is not enabled by default
	// +kubebuilder:default:HealthcheckDefault()
	Healthcheck HealthcheckSpec `json:"healthcheck,omitempty"`

	// If empty, autoscaling is not enabled by default
	// +kubebuilder:validation:Optional
	Autoscaling AutoscalingSpec `json:"autoscaling,omitempty"`

	// If empty, volumes won't be used
	// +kubebuilder:validation:Optional
	Deployment DeploymentSpec `json:"deployment,omitempty"`

	// Default value: size 1Gb
	// +kubebuilder:default:PersistentVolumeClaimDefault()
	PersistentVolumeClaim corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// ChallengeStatus defines the observed state of Challenge
type ChallengeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:Optional
	Status string `json:"challengeStatus"`
	// TODO: implement status for the challenges like READY and etc..
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Challenge is the Schema for the challenges API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=challenges,scope=Namespaced
type Challenge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChallengeSpec   `json:"spec,omitempty"`
	Status ChallengeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChallengeList contains a list of Challenge
type ChallengeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Challenge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Challenge{}, &ChallengeList{})
}
