// This file contains the reconcile function which is called when a CR is applied
// TODO: Change namespace and make operator watch all namespaces
// TODO; Check finalizers and add if necessary
package challenge

import (
	"context"

	kctfv1alpha1 "github.com/google/kctf/pkg/apis/kctf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName(name)

// Add creates a new Challenge Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileChallenge{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Challenge
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kctfv1alpha1.Challenge{},
	})

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileChallenge implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileChallenge{}

// ReconcileChallenge reconciles a Challenge object
type ReconcileChallenge struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Challenge object and makes changes based on the state read
// and what is in the Challenge.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileChallenge) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Challenge")

	// Fetch the Challenge instance
	challenge := &kctfv1alpha1.Challenge{}
	err := r.client.Get(context.TODO(), request.NamespacedName, challenge)
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

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: challenge.Name, Namespace: challenge.Namespace}, found)

	// Set default values not configured by kubebuilder
	SetDefaultValues(challenge)

	// Just enters here if it's a new deployment
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForChallenge(challenge)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)

		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}

		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil

	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Creates the service if it doesn't exist
	// Check existence of the service:
	serviceFound := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: challenge.Name, Namespace: challenge.Namespace}, serviceFound)

	// Just enter here if the service doesn't exist yet:
	if err != nil && errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		serv := r.serviceForChallenge(challenge)
		reqLogger.Info("Creating a new Service", "Service.Namespace", serv.Namespace, "Service.Name", serv.Name)
		err = r.client.Create(context.TODO(), serv)

		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Deployment.Service", serv.Namespace, "Service.Name", serv.Name)
			return reconcile.Result{}, err
		}

		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		reqLogger.Info("Deleting the Service", "Service.Namespace", serviceFound.Namespace, "Service.Name", serviceFound.Name)
		err = r.client.Delete(context.TODO(), serviceFound)

		if err != nil {
			reqLogger.Error(err, "Failed to erase Service", "Deployment.Service", serviceFound.Namespace, "Service.Name", serviceFound.Name)
			return reconcile.Result{}, err
		}

		// Service deleted successfully - return and requeue
		return reconcile.Result{}, err
	}

	// Ensure that the configurations in the CR are followed - Checks done everytime the CR is updated
	// change says if something in the configurations was different from what was found in the deploymeny
	change := UpdateConfigurations(challenge, found)

	// If there's a change it must requeue
	if change == true {
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
