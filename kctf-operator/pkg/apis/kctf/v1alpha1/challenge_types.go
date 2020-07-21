// +kubebuilder:validation:Required
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkSpec struct {

	// +kubebuilder:default:=true
	Public bool `json:"public,omitempty"`

	// +kubebuilder:default:=false
	Dns bool `json:"dns,omitempty"`

	// By default, one port is set with default values
	// +kubebuilder:validation:Optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`
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

	// kubebuilder:validation:Optional
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

	// +kubebuilder:validation:Optional
	Network NetworkSpec `json:"network,omitempty"`

	// If empty, healthcheck is not enabled by default
	// +kubebuilder:validation:Optional
	Healthcheck HealthcheckSpec `json:"healthcheck,omitempty"`

	// If empty, autoscaling is not enabled by default
	// +kubebuilder:validation:Optional
	Autoscaling AutoscalingSpec `json:"autoscaling,omitempty"`

	// If empty, volumes won't be used
	// +kubebuilder:validation:Optional
	Deployment DeploymentSpec `json:"deployment,omitempty"`

	// Default value: size 1Gb
	// +kubebuilder:validation:Optional
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
