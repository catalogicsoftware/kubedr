package metadatabackuppolicy

import (
	"context"

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

var log = logf.Log.WithName("controller_metadatabackuppolicy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MetadataBackupPolicy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMetadataBackupPolicy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("metadatabackuppolicy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MetadataBackupPolicy
	err = c.Watch(&source.Kind{Type: &kcxv1alpha1.MetadataBackupPolicy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

/*
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MetadataBackupPolicy
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kcxv1alpha1.MetadataBackupPolicy{},
	})
	if err != nil {
		return err
	}
*/

	return nil
}

// blank assignment to verify that ReconcileMetadataBackupPolicy implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMetadataBackupPolicy{}

// ReconcileMetadataBackupPolicy reconciles a MetadataBackupPolicy object
type ReconcileMetadataBackupPolicy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MetadataBackupPolicy object and makes changes based on the state read
// and what is in the MetadataBackupPolicy.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMetadataBackupPolicy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MetadataBackupPolicy")

	// Fetch the MetadataBackupPolicy instance
	instance := &kcxv1alpha1.MetadataBackupPolicy{}
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
	pod, err := r.metadataBackup(instance, request.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MetadataBackupPolicy instance as the owner and controller
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

func (r *ReconcileMetadataBackupPolicy) metadataBackup(cr *kcxv1alpha1.MetadataBackupPolicy, 
	namespace string) (*corev1.Pod, error) {

	backupLocation := &kcxv1alpha1.BackupLocation{}
	backupLocKey := types.NamespacedName{Namespace: namespace, Name: cr.Spec.Destination}
	err := r.client.Get(context.TODO(), backupLocKey, backupLocation)
	if err != nil {
		return nil, err
	}

	// TODO: Need to "Get" backup location resource and access data in it.
	s3EndPoint := "s3:" + backupLocation.Spec.Url + "/" + backupLocation.Spec.BucketName
	// reqLogger.Info("s3EndPoint: ", s3EndPoint)

	access_key := corev1.SecretKeySelector{}
	access_key.Name = backupLocation.Spec.Credentials
	access_key.Key = "access_key"

	secret_key := corev1.SecretKeySelector{}
	secret_key.Name = backupLocation.Spec.Credentials
	secret_key.Key = "secret_key"

	restic_password := corev1.SecretKeySelector{}
	restic_password.Name = backupLocation.Spec.Credentials
	restic_password.Key = "restic_repo_password"

	targetDirVolume := corev1.Volume{Name: "target-dir"}
	targetDirVolume.EmptyDir = &corev1.EmptyDirVolumeSource {}

	var t corev1.HostPathType = "Directory"
	certsDirVolume := corev1.Volume{Name: "certs-dir"}
	certsDirVolume.HostPath = &corev1.HostPathVolumeSource {
		Path: cr.Spec.CertsDir,
		Type: &t,
	}

	etcdCredsVolume := corev1.Volume{Name: "etcd-creds"}
	etcdCredsVolume.Secret = &corev1.SecretVolumeSource {
		SecretName: "etcd-creds",
	}


	return &corev1.Pod {
		ObjectMeta: metav1.ObjectMeta {
			Name:      cr.Name + "-backup-pod",
			Namespace: cr.Namespace,
		},

		Spec: corev1.PodSpec {
			RestartPolicy: "Never",
			HostNetwork: true,

			Volumes: []corev1.Volume {
				targetDirVolume,
				certsDirVolume,
				etcdCredsVolume,
			},

			InitContainers: []corev1.Container {
				{
					Name: cr.Name + "-etcd-snapshot",
					Image:   "k8s.gcr.io/etcd:3.3.15-0",

					VolumeMounts: []corev1.VolumeMount {
						{
							Name: "target-dir",
							MountPath: "/data",
						},
						{
							Name: "etcd-creds",
							MountPath: "/etcd_creds",
						},
					},

					Env: []corev1.EnvVar {
						{
							Name: "ETCDCTL_API",
							Value: "3",
						},
					},

					Command: []string{"etcdctl"},
					Args: []string {
						"--endpoints=" + cr.Spec.EtcdEndpoint,
						"--cacert=/etcd_creds/ca.crt",
						"--cert=/etcd_creds/client.crt",
						"--key=/etcd_creds/client.key",
						"--debug", "snapshot", "save",
						"/data/etcd-snapshot.db",
					},
				},

				{
					Name: cr.Name + "-certificates-copy",
					Image:   "busybox",

					VolumeMounts: []corev1.VolumeMount {
						{
							Name: "target-dir",
							MountPath: "/data",
						},
						{
							Name: "certs-dir",
							MountPath: "/certs_dir",
						},
					},

					Command: []string{"/bin/sh"},
					Args: []string {
						"-c",
						"mkdir -p /data/certificates && cp -R /certs_dir/* /data/certificates",
					},
				},
			},

			Containers: []corev1.Container {
				{
					Name: cr.Name + "-kcx-backup",
					Image:   "restic/restic",

					VolumeMounts: []corev1.VolumeMount {
						{
							Name: "target-dir",
							MountPath: "/data",
						},
					},

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

					Args: []string {
						"-r", s3EndPoint, "--verbose",
						"backup", "/data",
					},
				},
			},
		},
	}, nil
}
