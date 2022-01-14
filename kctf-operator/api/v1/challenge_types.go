/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

type PortSpec struct {
	// Name of the port
	Name string `json:"name,omitempty"`

	// TargetPort is not optional
	// +kubebuilder:validation:Required
	TargetPort intstr.IntOrString `json:"targetPort"`

	// Port
	Port int32 `json:"port,omitempty"`

	// Protocol is not optional
	// +kubebuilder:validation:Required
	Protocol corev1.Protocol `json:"protocol"`

	// Extra domains for managed certificates. Only used for type HTTPS.
	Domains []string `json:"domains,omitempty"`
}

// Network specifications for the service
type NetworkSpec struct {

	// +kubebuilder:default:=false
	Public bool `json:"public,omitempty"`

	// By default, one port is set with default values
	Ports []PortSpec `json:"ports,omitempty"`
}

// Healthcheck specifications
type HealthcheckSpec struct {

	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// Image for the healthcheck container
	// +kubebuilder:default:="healthcheck"
	Image string `json:"image,omitempty"`
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
	// +kubebuilder:validation:Required
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty" protobuf:"varint,4,opt,name=targetCPUUtilizationPercentage"`
}

// ChallengeSpec defines the desired state of Challenge
type ChallengeSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Image used by the deployment
	// +kubebuilder:default:="challenge"
	Image string `json:"image"`

	// Shows if the challenge is ready to be deployed, if not,
	// it sets the replicas to 0 and disables services/ingress
	// +kubebuilder:default:=false
	Deployed bool `json:"deployed,omitempty"`

	// The desired quantity of replicas if horizontal pod autoscaler is disabled
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// The quantity of seconds of the proof of work
	// +kubebuilder:default:=0
	PowDifficultySeconds int `json:"powDifficultySeconds,omitempty"`

	// The network specifications: if it's public or not and specifications about ports
	Network NetworkSpec `json:"network,omitempty"`

	// Healthcheck checks if the challenge works
	// If empty, healthcheck is not enabled by default
	// +kubebuilder:validation:Required
	Healthcheck HealthcheckSpec `json:"healthcheck,omitempty"`

	// Autoscaling features determine quantity of replicas and CPU utilization
	// If empty, autoscaling is not enabled by default
	HorizontalPodAutoscalerSpec *HorizontalPodAutoscalerSpec `json:"horizontalPodAutoscalerSpec,omitempty"`

	// PodTemplate is used to set the template for the deployment's pod,
	// so that an author can add volumeMounts and other extra features
	PodTemplate *corev1.PodTemplate `json:"podTemplate,omitempty"`

	// Names of the desired PersistentVolumeClaims
	PersistentVolumeClaims []string `json:"persistentVolumeClaims,omitempty"`

	AllowConnectTo []string `json:"allowConnectTo,omitempty"`
}

// ChallengeStatus defines the observed state of Challenge
type ChallengeStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Says if the challenge is up to date or being updated
	// +kubebuilder:default:="up-to-date"
	Status corev1.PodPhase `json:"status"`

	// Shows healthcheck returns
	// +kubebuilder:default:="disabled"
	Health string `json:"health"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Challenge is the Schema for the challenges API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=challenges,scope=Namespaced
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Status", type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Deployed",type=boolean,JSONPath=`.spec.deployed`
// +kubebuilder:printcolumn:name="Public",type=boolean,JSONPath=`.spec.network.public`
type Challenge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChallengeSpec   `json:"spec,omitempty"`
	Status ChallengeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChallengeList contains a list of Challenge
type ChallengeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Challenge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Challenge{}, &ChallengeList{})
}
