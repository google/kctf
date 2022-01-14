module github.com/google/kctf

go 1.16

require (
	github.com/GoogleCloudPlatform/gke-managed-certs v1.0.5
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/ingress-gce v1.14.5
	sigs.k8s.io/controller-runtime v0.10.0
)

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210323165736-1a6458611d18
