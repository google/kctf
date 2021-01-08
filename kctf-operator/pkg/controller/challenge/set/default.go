// File that set values correctly and return default values that weren't specified
package set

import (
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Function to return the default ports
func portsDefault() []kctfv1alpha1.PortSpec {
	var portsDefault = []kctfv1alpha1.PortSpec{
		kctfv1alpha1.PortSpec{
			// Keeping the same name as in previous network file
			Name:       "netcat",
			Port:       1337,
			TargetPort: intstr.FromInt(1337),
			Protocol:   "TCP",
		},
	}
	return portsDefault
}

// Function to check if all is set to default
func DefaultValues(challenge *kctfv1alpha1.Challenge, scheme *runtime.Scheme) {
	// Sets default ports
	if challenge.Spec.Network.Ports == nil {
		challenge.Spec.Network.Ports = portsDefault()
	}
}
