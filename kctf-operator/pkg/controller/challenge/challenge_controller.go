// This file contains the reconcile function which is called when a CR is applied
package challenge

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	"github.com/google/kctf/pkg/controller/challenge/autoscaling"
	"github.com/google/kctf/pkg/controller/challenge/deployment"
	"github.com/google/kctf/pkg/controller/challenge/dns"
	"github.com/google/kctf/pkg/controller/challenge/pow"
	"github.com/google/kctf/pkg/controller/challenge/secrets"
	"github.com/google/kctf/pkg/controller/challenge/service"
	"github.com/google/kctf/pkg/controller/challenge/set"
	"github.com/google/kctf/pkg/controller/challenge/status"
	"github.com/google/kctf/pkg/controller/challenge/volumes"
	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const name = "challenge-controller"

// Add creates a new Challenge Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r := newReconciler(mgr)
	err := add(mgr, r)
	return err
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileChallenge{client: mgr.GetClient(), scheme: mgr.GetScheme(), log: logf.Log.WithName(name)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(name, mgr, controller.Options{Reconciler: r})

	if err != nil {
		return err
	}

	// Watch for changes to primary resource Challenge
	err = c.Watch(&source.Kind{Type: &kctfv1alpha1.Challenge{}}, &handler.EnqueueRequestForObject{})

	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Challenge
	objs := []runtime.Object{&corev1.Pod{}, &appsv1.Deployment{}, &autoscalingv1.HorizontalPodAutoscaler{},
		&corev1.Service{}, &netv1beta1.Ingress{}, &corev1.PersistentVolumeClaim{}, &corev1.PersistentVolume{},
		&corev1.ConfigMap{}, &corev1.Secret{}}

	for _, obj := range objs {
		err = c.Watch(&source.Kind{Type: obj}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &kctfv1alpha1.Challenge{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileChallenge reconciles a Challenge object
type ReconcileChallenge struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

// Reconcile reads that state of the cluster for a Challenge object and makes changes based on the state read
// and what is in the Challenge.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

func (r *ReconcileChallenge) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	reqLogger := r.log.WithValues("Challenge ", request.Name, " with namespace ", request.Namespace)
	reqLogger.Info("Reconciling Challenge")

	// Fetch the Challenge instance
	challenge := &kctfv1alpha1.Challenge{}
	requeue, err := r.fetchChallenge(challenge, request, ctx)
	if err != nil || requeue {
		return reconcile.Result{}, err
	}

	// Checks if namespace is acceptable and if not, it deletes the challenge
	if !isNamespaceAcceptable(request.NamespacedName) {
		reqLogger.Info("Can't accept namespace different from name of the challenge. Please change namespace.",
			request.Name, " with namespace ", request.Namespace)
		reqLogger.Info("Deleting challenge")
		err = r.client.Delete(ctx, challenge)
		if err != nil {
			status.Update(false, err, challenge, r.client,
				r.log, ctx)
			log.Error(err, "Failed to delete challenge")
		}
		return reconcile.Result{}, err
	}

	// Set default values not configured by kubebuilder
	set.DefaultValues(challenge, r.scheme)

	// Ensure that the configurations in the CR are followed - Checks done everytime the CR is updated
	// change says if something in the configurations was different from what was found in the deployment
	requeue, err = updateConfigurations(challenge, r.client, r.scheme, r.log, ctx)
	status.Update(requeue, err, challenge, r.client, r.log, ctx)

	if err != nil {
		reqLogger.Error(err, "Failed to update Challenge")
		return reconcile.Result{}, err
	} else if requeue == true {
		reqLogger.Info("Challenge updated successfully", "Name: ",
			request.Name, " with namespace ", request.Namespace)
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileChallenge) fetchChallenge(challenge *kctfv1alpha1.Challenge,
	request reconcile.Request, ctx context.Context) (bool, error) {
	err := r.client.Get(ctx, request.NamespacedName, challenge)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.log.Info("Challenge resource not found. Ignoring since object must be deleted", "Name: ",
				request.Name, " with namespace ", request.Namespace)
			return true, nil
		}

		// Error reading the object - requeue the request.
		r.log.Error(err, "Failed to get Challenge")
		return true, err
	}

	return false, nil
}

// Function that returns if the chosen namespace is acceptable or no to prevent errors
func isNamespaceAcceptable(namespacedName types.NamespacedName) bool {
	if namespacedName.Name != namespacedName.Namespace ||
		namespacedName.Namespace == "default" || namespacedName.Namespace == "kube-system" {
		return false
	}
	return true
}

func updateConfigurations(challenge *kctfv1alpha1.Challenge, cl client.Client, scheme *runtime.Scheme,
	log logr.Logger, ctx context.Context) (bool, error) {
	// We check if there's an error in each update
	updateFunctions := []func(challenge *kctfv1alpha1.Challenge, client client.Client, scheme *runtime.Scheme,
		log logr.Logger, ctx context.Context) (bool, error){volumes.Update,
		pow.Update, secrets.Update, deployment.Update, service.Update, dns.Update,
		autoscaling.Update}

	for _, updateFunction := range updateFunctions {
		requeue, err := updateFunction(challenge, cl, scheme, log, ctx)
		if err != nil {
			return requeue, err
		}
	}

	return false, nil
}
