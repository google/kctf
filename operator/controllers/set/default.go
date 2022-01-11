// File that set values correctly and return default values that weren't specified
package set

import (
	kctfv1 "github.com/google/kctf/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Function to return the default ports
func portsDefault() []kctfv1.PortSpec {
	var portsDefault = []kctfv1.PortSpec{
		kctfv1.PortSpec{
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
func DefaultValues(challenge *kctfv1.Challenge, scheme *runtime.Scheme) {
	// Sets default ports
	if challenge.Spec.Network.Ports == nil {
		challenge.Spec.Network.Ports = portsDefault()
	}
}
