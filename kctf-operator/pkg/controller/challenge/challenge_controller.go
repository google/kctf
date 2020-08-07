// This file contains the reconcile function which is called when a CR is applied
// TODO: Synthesize this file : create a function that is called multiple times
// since there's a lot of code that is repeated
// TODO: change deletion and creation of service/ingress/etc to challenge_update
package challenge

import (
	"context"

	"github.com/go-logr/logr"
	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
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

//var log = logf.Log.WithName(name)

// Add creates a new Challenge Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileChallenge{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// TODO: discover why deleting the challenge isn't deleting the service, the ingress and the autoscaling
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
		&corev1.Service{}, &netv1beta1.Ingress{}}

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
	r.log = logf.Log.WithName(name)

	reqLogger := r.log.WithValues("Challenge ", request.Name, " with namespace ", request.Namespace)
	reqLogger.Info("Reconciling Challenge")

	// Fetch the Challenge instance
	challenge := &kctfv1alpha1.Challenge{}
	err := r.client.Get(ctx, request.NamespacedName, challenge)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Challenge resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Challenge")
		return reconcile.Result{}, err
	}

	// Set default values not configured by kubebuilder
	SetDefaultValues(challenge)

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, found)

	// Just enters here if it's a new deployment
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		return r.CreateDeployment(challenge, ctx)

	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Creates autoscaling object
	// Checks if an autoscaling was configured
	// If enabled, it checks if it already exists
	autoscalingFound := &autoscalingv1.HorizontalPodAutoscaler{}
	err = r.client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, autoscalingFound)

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil && err != nil && errors.IsNotFound(err) {
		// creates autoscaling if it doesn't exist yet
		return r.CreateAutoscaling(challenge, ctx)
	}

	if challenge.Spec.HorizontalPodAutoscalerSpec == nil && err == nil {
		// delete autoscaling
		return r.DeleteAutoscaling(autoscalingFound, ctx)
	}

	// Creates the service if it doesn't exist
	// Check existence of the service:
	serviceFound := &corev1.Service{}
	ingressFound := &netv1beta1.Ingress{}
	err = r.client.Get(ctx, types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, serviceFound)
	err_ingress := r.client.Get(ctx, types.NamespacedName{Name: "https",
		Namespace: challenge.Namespace}, ingressFound)

	// Just enter here if the service doesn't exist yet:
	if err != nil && errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		return r.CreateServiceAndIngress(challenge, ctx, err_ingress)

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		return r.DeleteServiceAndIngress(serviceFound, ingressFound, ctx, err_ingress)
	}

	// Ensure that the configurations in the CR are followed - Checks done everytime the CR is updated
	// change says if something in the configurations was different from what was found in the deployment
	change := UpdateConfigurations(challenge, found)

	// If there's a change it must requeue
	if change == true {
		err = r.client.Update(ctx, found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment",
				"Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
