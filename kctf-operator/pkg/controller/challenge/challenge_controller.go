// This file contains the reconcile function which is called when a CR is applied
// TODO: Change namespace
// TODO: Synthesize this file : create a function that is called multiple times
// since there's a lot of code that is repeated
// TODO; Check finalizers and add if necessary
package challenge

import (
	"context"

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
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, found)

	// Set default values not configured by kubebuilder
	SetDefaultValues(challenge)

	// Just enters here if it's a new deployment
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForChallenge(challenge)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace",
			dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)

		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
				dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}

		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil

	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Creates autoscaling object
	// Checks if an autoscaling was configured
	// If enabled, it checks if it already exists
	autoscalingFound := &autoscalingv1.HorizontalPodAutoscaler{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, autoscalingFound)

	if challenge.Spec.HorizontalPodAutoscalerSpec != nil && err != nil && errors.IsNotFound(err) {
		// creates autoscaling if it doesn't exist yet
		autoscaling := r.autoscalingForChallenge(challenge)
		reqLogger.Info("Creating a Autoscaling")
		err = r.client.Create(context.TODO(), autoscaling)

		if err != nil {
			reqLogger.Error(err, "Failed to create Autoscaling")
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	}

	// TODO: pass this and other deletes to challenge_update
	if challenge.Spec.HorizontalPodAutoscalerSpec == nil && err == nil {
		// delete autoscaling it's false
		reqLogger.Info("Deleting Autoscaling")
		err = r.client.Delete(context.TODO(), autoscalingFound)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Autoscaling")
			return reconcile.Result{}, err
		}
	}

	// Creates the service if it doesn't exist
	// Check existence of the service:
	serviceFound := &corev1.Service{}
	ingressFound := &netv1beta1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: challenge.Name,
		Namespace: challenge.Namespace}, serviceFound)
	err_ingress := r.client.Get(context.TODO(), types.NamespacedName{Name: "https",
		Namespace: challenge.Namespace}, ingressFound)

	// Just enter here if the service doesn't exist yet:
	if err != nil && errors.IsNotFound(err) && challenge.Spec.Network.Public == true {
		// Define a new service if the challenge is public
		serv, ingress := r.serviceForChallenge(challenge)
		// See if there's any port defined for the service
		reqLogger.Info("Creating a new Service", "Service.Namespace",
			serv.Namespace, "Service.Name", serv.Name)
		err = r.client.Create(context.TODO(), serv)

		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace",
				serv.Namespace, "Service.Name", serv.Name)
			return reconcile.Result{}, err
		}

		// Create ingress, if there's any https
		if err_ingress != nil && errors.IsNotFound(err_ingress) {
			// If there's a port HTTPS
			if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == true {
				// Create ingress in the client
				reqLogger.Info("Creating a new Ingress", "Ingress.Namespace", ingress.Namespace,
					"Ingress.Name", ingress.Name)
				err = r.client.Create(context.TODO(), ingress)

				if err != nil {
					reqLogger.Error(err, "Failed to create new Ingress", "Ingress.Namespace", ingress.Namespace,
						"Ingress.Name", ingress.Name)
					return reconcile.Result{}, err
				}

				// Ingress created successfully
				return reconcile.Result{}, err
			}

			if ingress.Spec.Backend != nil && challenge.Spec.Network.Dns == false {
				reqLogger.Info("Failed to create Ingress instance, DNS isn't enabled. Challenge won't be reconciled here.")
			}
		}

		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil

		// When service exists and public is changed to false
	} else if err == nil && challenge.Spec.Network.Public == false {
		reqLogger.Info("Deleting the Service", "Service.Namespace", serviceFound.Namespace,
			"Service.Name", serviceFound.Name)
		err = r.client.Delete(context.TODO(), serviceFound)

		if err != nil {
			reqLogger.Error(err, "Failed to delete Service", "Service.Namespace", serviceFound.Namespace,
				"Service.Name", serviceFound.Name)
			return reconcile.Result{}, err
		}

		// Delete ingress if existent
		if err_ingress == nil {
			reqLogger.Info("Deleting the Ingress", "Ingress.Namespace", ingressFound.Namespace, "Ingress.Name", ingressFound.Name)
			err = r.client.Delete(context.TODO(), ingressFound)

			if err != nil {
				reqLogger.Error(err, "Failed to delete Ingress", "Ingress.Namespace", ingressFound.Namespace,
					"Ingress.Name", ingressFound.Name)
				return reconcile.Result{}, err
			}
		}

		// Service deleted successfully - return and requeue
		return reconcile.Result{}, err
	}

	// Ensure that the configurations in the CR are followed - Checks done everytime the CR is updated
	// change says if something in the configurations was different from what was found in the deployment
	change := UpdateConfigurations(challenge, found)

	// If there's a change it must requeue
	if change == true {
		err = r.client.Update(context.TODO(), found)
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
