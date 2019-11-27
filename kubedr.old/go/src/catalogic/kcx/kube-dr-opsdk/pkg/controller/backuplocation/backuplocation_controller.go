package backuplocation

import (
	"context"
//	"time"
//	"math/rand"

	kcxv1alpha1 "catalogic/kcx/kube-dr/pkg/apis/kcx/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_backuplocation")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BackupLocation Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBackupLocation{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("backuplocation-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BackupLocation
	err = c.Watch(&source.Kind{Type: &kcxv1alpha1.BackupLocation{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/* Don't monitor the pods created as a result of creating backup location. 
* Otherwise, when the pod is deleted, the controller is creating the pod again. 
*/

/*
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner BackupLocation
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kcxv1alpha1.BackupLocation{},
	})
	if err != nil {
		return err
	}
*/

	return nil
}

// blank assignment to verify that ReconcileBackupLocation implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBackupLocation{}

// ReconcileBackupLocation reconciles a BackupLocation object
type ReconcileBackupLocation struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BackupLocation object and makes changes based on the state read
// and what is in the BackupLocation.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBackupLocation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling BackupLocation")

	// Fetch the BackupLocation instance
	instance := &kcxv1alpha1.BackupLocation{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := initializeResticRepo(instance)

	// Set BackupLocation instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
/*
func newPodForCR(cr *kcxv1alpha1.BackupLocation) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
*/

func initializeResticRepo(cr *kcxv1alpha1.BackupLocation) *corev1.Pod {
	// id := time.Now().Unix() + rand.Int63()
	s3EndPoint := "s3:" + cr.Spec.Url + "/" + cr.Spec.BucketName
	// reqLogger.Info("s3EndPoint: ", s3EndPoint)

	labels := map[string]string{
		"app": cr.Name,
	}

	access_key := corev1.SecretKeySelector{}
	access_key.Name = cr.Spec.Credentials
	access_key.Key = "access_key"

	secret_key := corev1.SecretKeySelector{}
	secret_key.Name = cr.Spec.Credentials
	secret_key.Key = "secret_key"

	restic_password := corev1.SecretKeySelector{}
	restic_password.Name = cr.Spec.Credentials
	restic_password.Key = "restic_repo_password"

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta {
			Name:      cr.Name + "-init-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec {
			Containers: []corev1.Container {
				{
					Name: cr.Name + "-init",
					Image:   "restic/restic",
					Args: []string{"-r", s3EndPoint, "init"},
					Env: []corev1.EnvVar {
						{
							Name: "MINIO_ACCESS_KEY",
							ValueFrom: &corev1.EnvVarSource {
								SecretKeyRef: &access_key,
							},
						},
						{
							Name: "MINIO_SECRET_KEY",
							ValueFrom: &corev1.EnvVarSource {
								SecretKeyRef: &secret_key,
							},
						},
						{
							Name: "RESTIC_PASSWORD",
							ValueFrom: &corev1.EnvVarSource {
								SecretKeyRef: &restic_password,
							},
						},
					},
				},
			},
			RestartPolicy: "Never",
		},
	}
}
