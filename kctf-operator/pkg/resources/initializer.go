package resources

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log logr.Logger = logf.Log.WithName("cmd")

func InitializeOperator(client *client.Client) error {
	// Creates the objects that enable the DNS, external DNS and etc

	// Generate keys
	generateKeys()

	objectFunctions := []func() runtime.Object{NewApparmorProfiles, NewAllowDns, NewServiceAccountGcsFuseSa,
		NewServiceAccountExternalDnsSa, NewExternalDnsClusterRole, NewExternalDnsClusterRoleBinding,
		NewExternalDnsDeployment, NewDaemonSetCtf, NewDaemonSetGcsFuse, NewNetworkPolicyBlockInternal,
		NewNetworkPolicyKubeSystem, NewSecretPowBypass, NewSecretPowBypassPub}

	names := []string{"Apparmor Profiles", "Service Account Gcs Fuse", "Service Account External DNS",
		"Allow DNS", "External DNS Cluster Role",
		"External DNS Cluster Role Binding", "External DNS Deployment", "Daemon Set Ctf", "Daemon Set Gcs Fuse",
		"Network Policy Block Internal", "Network Policy Kube System", "Secret for PowBypass", "Secret for PowBypassPub"}

	for i, newObject := range objectFunctions {

		obj := newObject()

		// Creates the object
		err := (*client).Create(context.Background(), obj)

		// Checks if the error is already exists, because if it is, it's not a problem
		if err != nil {
			if errors.IsAlreadyExists(err) {
				log.Info("This object already exists.", "Name: ", names[i])
			} else {
				log.Error(err, names[i])
				log.Info(names[i])
			}
		} else {
			log.Info("Created object.", "Name:", names[i])
		}
	}

	return nil
}
