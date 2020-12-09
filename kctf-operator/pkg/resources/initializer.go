package resources

import (
	"context"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log logr.Logger = logf.Log.WithName("cmd")

func InitializeOperator(client *client.Client) error {
	// Creates the objects that enable the DNS, external DNS and etc

	objectFunctions := []func() runtime.Object{NewApparmorProfiles, NewServiceAccountGcsFuseSa,
		NewServiceAccountExternalDnsSa, NewExternalDnsClusterRole, NewExternalDnsClusterRoleBinding,
		NewExternalDnsDeployment, NewDaemonSetCtf, NewDaemonSetGcsFuse, NewSecretPowBypass,
		NewSecretPowBypassPub, NewNetworkPolicyBlockInternal, NewAllowDns}

	names := []string{"Apparmor Profiles", "Service Account Gcs Fuse", "Service Account External DNS",
		"External DNS Cluster Role", "External DNS Cluster Role Binding", "External DNS Deployment",
		"Daemon Set Ctf", "Daemon Set Gcs Fuse", "Secret for PowBypass", "Secret for PowBypassPub",
		"Network Policy Block Internal", "Allow DNS"}

	for i, newObject := range objectFunctions {

		obj := newObject()

		// Creates the object
		err := (*client).Create(context.Background(), obj)

		// Checks if the error is already exists, because if it is, it's not a problem
		if err != nil {
			if errors.IsAlreadyExists(err) {
				log.Info("This object already exists.", "Name: ", names[i])
				// Try to update the resource instead
				err = (*client).Update(context.Background(), obj)
			}
			if err != nil {
				log.Error(err, names[i])
				log.Info(names[i])
				return err
			}
		} else {
			log.Info("Created object.", "Name:", names[i])
		}
	}

	f, err := os.Create("/tmp/initialized")
	if err != nil {
		log.Error(err, "Could not create file for ReadinessProbe")
		return err
	}
	f.Close()

	return nil
}
