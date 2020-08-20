// This file is responsible for generating CRD (Custom Resource Definition)
// kubebuilder might be used to set: default values, optional fields and etc
// +kubebuilder:validation:Required
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

type PortSpec struct {
	//+optional
	Name string `json:"name"`

	// TargetPort is not optional
	TargetPort intstr.IntOrString `json:"targetPort"`

	//+optional
	Port int32 `json:"port"`

	// Protocol is not optional
	Protocol corev1.Protocol `json:"protocol"`
}

// Network specifications for the service
type NetworkSpec struct {

	// +kubebuilder:default:=false
	Public bool `json:"public,omitempty"`

	// +kubebuilder:default:=false
	Dns bool `json:"dns,omitempty"`

	// By default, one port is set with default values
	// +optional
	Ports []PortSpec `json:"ports,omitempty"`
}

// Healthcheck specifications
type HealthcheckSpec struct {

	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:default:="healthcheck"
	Container string `json:"container,omitempty"`
}

// HorizontalPodAutoscalerSpec without ScaleTargetRef
type HorizontalPodAutoscalerSpec struct {
	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	// +optional
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty" protobuf:"varint,4,opt,name=targetCPUUtilizationPercentage"`
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
	// +optional
	Network NetworkSpec `json:"network,omitempty"`

	// Healthcheck checks if the challenge works
	// If empty, healthcheck is not enabled by default
	// +optional
	Healthcheck HealthcheckSpec `json:"healthcheck,omitempty"`

	// Autoscaling features determine quantity of replicas and CPU utilization
	// If empty, autoscaling is not enabled by default
	// +optional
	HorizontalPodAutoscalerSpec *HorizontalPodAutoscalerSpec `json:"horizontalPodAutoscalerSpec,omitempty"`

	// PodTemplate is used to set the template for the deployment's pod,
	// so that an author can add volumeMounts and other extra features
	// +optional
	PodTemplate *corev1.PodTemplate `json:"podTemplate"`

	// PersistentVolumeClaim are used to determine how much resources the author requires for its challenge
	// Default value: size 1Gb
	// +optional
	Claims []string `json:"claims,omitempty"`
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
