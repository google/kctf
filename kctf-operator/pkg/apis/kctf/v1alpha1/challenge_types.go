// This file is responsible for generating CRD (Custom Resource Definition)
// kubebuilder might be used to set: default values, optional fields and etc
// +kubebuilder:validation:Required
package v1alpha1

import (
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

type PortSpec struct {
	//+kubebuilder:validation:Optional
	Name string `json:"name"`

	// TargetPort is not optional
	TargetPort intstr.IntOrString `json:"targetPort"`

	//+kubebuilder:validation:Optional
	Port int32 `json:"port"`

	// Protocol is not optional
	Protocol corev1.Protocol `json:"protocol"`
}

// Network specifications for the service
type NetworkSpec struct {

	// +kubebuilder:default:=true
	Public bool `json:"public,omitempty"`

	// +kubebuilder:default:=false
	Dns bool `json:"dns,omitempty"`

	// By default, one port is set with default values
	// +kubebuilder:validation:Optional
	Ports []PortSpec `json:"ports,omitempty"`
}

// Healthcheck specifications
type HealthcheckSpec struct {

	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:default:="healthcheck"
	Container string `json:"container,omitempty"`
}

// Autoscaling Specifications
type AutoscalingSpec struct {

	// Minimum quantity of replicas
	// +kubebuilder:default:=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// Maximum quantity of replicas
	// +kubebuilder:default:=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// Target of CPU utilizantion in percentage
	// If empty, this feature won't be used
	// +kubebuilder:default:=50
	TargetCPUUtilizationPercentage int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}

// ChallengeSpec defines the desired state of Challenge
type ChallengeSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Image used by the deployment
	// Not optional and image should be passed by user (by now)
	ImageTemplate string `json:"imageTemplate"`

	// Shows if the challenge is ready to be deployed, if not,
	// it sets the replicas to 0
	// +kubebuilder:default:=false
	Deployed bool `json:"deployed,omitempty"`

	// The quantity of seconds of the proof of work
	// +kubebuilder:default:=0
	PowDifficultySeconds int32 `json:"powDifficultySeconds,omitempty"`

	// The network specifications: if it's public or not, if it uses dns or not and specifications about ports
	// +kubebuilder:validation:Optional
	Network NetworkSpec `json:"network,omitempty"`

	// Healthcheck checks if the challenge works
	// If empty, healthcheck is not enabled by default
	// +kubebuilder:validation:Optional
	Healthcheck HealthcheckSpec `json:"healthcheck,omitempty"`

	// Autoscaling features determine quantity of replicas and CPU utilization
	// If empty, autoscaling is not enabled by default
	// +kubebuilder:validation:Optional
	HorizontalAutoscaling autoscalingv1.HorizontalPodAutoscaler `json:"horizontalAutoscaling,omitempty"`

	// PodTemplate is used to set the paths of sessions and uploads
	// If empty, volumes won't be used
	// +kubebuilder:validation:Optional
	PodTemplate corev1.PodTemplate `json:"podTemplate,omitempty"`

	// PersistentVolumeClaim are used to determine how much resources the author requires for its challenge
	// Default value: size 1Gb
	// +kubebuilder:validation:Optional
	PersistentVolumeClaims corev1.PersistentVolumeClaimList `json:"persistentVolumeClaim,omitempty"`
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
