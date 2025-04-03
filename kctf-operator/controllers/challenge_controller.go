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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	kctfv1 "github.com/google/kctf/api/v1"

	"github.com/google/kctf/controllers/autoscaling"
	"github.com/google/kctf/controllers/deployment"
	"github.com/google/kctf/controllers/network-policy"
	"github.com/google/kctf/controllers/pow"
	"github.com/google/kctf/controllers/secrets"
	"github.com/google/kctf/controllers/service"
	"github.com/google/kctf/controllers/set"
	"github.com/google/kctf/controllers/status"
	"github.com/google/kctf/controllers/volumes"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// ChallengeReconciler reconciles a Challenge object
type ChallengeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kctf.dev,resources=challenges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kctf.dev,resources=challenges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kctf.dev,resources=challenges/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud.google.com,resources=backendconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.gke.io,resources=managedcertificates,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ChallengeReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx).WithValues("Challenge ", request.Name, " with namespace ", request.Namespace)
	reqLogger.Info("Reconciling Challenge")

	// Fetch the Challenge instance
	challenge := &kctfv1.Challenge{}
	requeue, err := r.fetchChallenge(challenge, request, reqLogger, ctx)
	if err != nil || requeue {
		return ctrl.Result{}, err
	}

	// Set default values not configured by kubebuilder
	set.DefaultValues(challenge, r.Scheme)

	// Ensure that the configurations in the CR are followed - Checks done everytime the CR is updated
	// change says if something in the configurations was different from what was found in the deployment
	requeue, err = updateConfigurations(challenge, r.Client, r.Scheme, reqLogger, ctx)
	status.Update(requeue, err, challenge, r.Client, reqLogger, ctx)

	if err != nil {
		reqLogger.Error(err, "Failed to update Challenge")
		return ctrl.Result{}, err
	} else if requeue == true {
		reqLogger.Info("Challenge updated successfully", "Name: ",
			request.Name, " with namespace ", request.Namespace)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChallengeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kctfv1.Challenge{}).
		Owns(&appsv1.Deployment{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Owns(&corev1.Service{}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.PersistentVolume{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []ctrl.Request {
				if a.GetNamespace() == "kctf-system" {
					challengeList := &kctfv1.ChallengeList{}
					err := mgr.GetClient().List(context.Background(), challengeList)
					if err != nil {
						// log.Error(err, "Failed to obtain a list of all challenges for updating a secret")
						return nil
					}
					requestList := []ctrl.Request{}
					for i := range challengeList.Items {
						requestList = append(requestList, ctrl.Request{
							NamespacedName: types.NamespacedName{
								Name:      challengeList.Items[i].Name,
								Namespace: challengeList.Items[i].Namespace,
							}})
					}
					return requestList
				}
				return nil
			})).
		Complete(r)
}

func (r *ChallengeReconciler) fetchChallenge(challenge *kctfv1.Challenge,
	request ctrl.Request, log logr.Logger, ctx context.Context) (bool, error) {
	err := r.Client.Get(ctx, request.NamespacedName, challenge)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Challenge resource not found. Ignoring since object must be deleted", "Name: ",
				request.Name, " with namespace ", request.Namespace)
			return true, nil
		}

		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Challenge")
		return true, err
	}

	return false, nil
}

func updateConfigurations(challenge *kctfv1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// We check if there's an error in each update
	updateFunctions := []func(challenge *kctfv1.Challenge, client client.Client, scheme *runtime.Scheme,
		log logr.Logger, ctx context.Context) (bool, error){network.Update, volumes.Update,
		pow.Update, secrets.Update, deployment.Update, service.Update, autoscaling.Update}

	for _, updateFunction := range updateFunctions {
		requeue, err := updateFunction(challenge, cl, scheme, log, ctx)
		if err != nil {
			return requeue, err
		}
	}

	return false, nil
}
