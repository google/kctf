module github.com/google/kctf

go 1.13

require (
	github.com/GoogleCloudPlatform/gke-managed-certs v1.0.1
	github.com/go-logr/logr v0.1.0
	github.com/operator-framework/operator-sdk v0.19.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	k8s.io/api => k8s.io/api v0.18.9
	k8s.io/client-go => k8s.io/client-go v0.17.2 // Required by prometheus-operator
)
