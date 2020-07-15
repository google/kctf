package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PortSpec struct {
	// TODO: port shouldn't be an obligatory field
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Protocol   string `json:"protocol"`
}

type NetworkSpec struct {
	Public bool       `json:"public"`
	Dns    bool       `json:"dns"`
	Ports  []PortSpec `json:"ports"`
}

type HealthcheckSpec struct {
	Enabled bool `json:"enabled"`
	// TODO: container should be always healthcheck, i guess
	Container string `json:"container"`
}

type AutoscalingSpec struct {
	Enabled                        bool `json:"enabled"`
	MinReplicas                    int  `json:"minReplicas"`
	MaxReplicas                    int  `json:"maxReplicas"`
	TargetCPUUtilizationPercentage int  `json:"targetCPUUtilizationPercentage"`
}

type DeploymentSpec struct {
	PersistentVolumeClaim corev1.PersistentVolumeClaim `json:"persistentVolumeClaim"`
	Template              corev1.PodTemplate           `json:"podTemplate"`
}

// ChallengeSpec defines the desired state of Challenge
type ChallengeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// TODO: add other fields
	// TODO: add maximum, minimum and default value
	// TODO: discover how to include disk stuff
	// TODO: ready in status or in spec?
	ImageTemplate        string          `json:"imageTemplate"`
	PowDifficultySeconds int             `json:"powDifficultySeconds"`
	Network              NetworkSpec     `json:"network"`
	Healthcheck          HealthcheckSpec `json:"healthcheck"`
	Autoscaling          AutoscalingSpec `json:"autoscaling"`
	Deployment           DeploymentSpec  `json:"deployment"`
}

// ChallengeStatus defines the observed state of Challenge
type ChallengeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
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
