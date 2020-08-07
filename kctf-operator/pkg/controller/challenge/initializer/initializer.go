package initializer

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitializeOperator(client *client.Client) error {
	// Creates the objects that enable the dns, external dns and etc
	objectFunctions := []func() runtime.Object{NewApparmorProfiles, NewAllowDns, NewExternalDnsClusterRole,
		NewExternalDnsClusterRoleBinding, NewExternalDnsDeployment, NewDaemonSetCtf, NewDaemonSetGcsFuse,
		NewNetworkPolicyBlockInternal, NewNetworkPolicyKubeSystem}

	// TODO: add garbage collection back for testing!
	for _, newObject := range objectFunctions {
		// Creates an object and checks if it was correctly created
		obj := newObject()

		err := (*client).Create(context.Background(), obj)

		if err != nil {
			return err
		}
	}

	return nil
}
