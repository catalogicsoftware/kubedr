/*
Copyright 2020 Catalogic Software

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
	"fmt"
	"os"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubedrv1alpha1 "kubedr/api/v1alpha1"
)

// BackupLocationReconciler reconciles a BackupLocation object
type BackupLocationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=backuplocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=backuplocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=create;get
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;get;list;watch;update

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.

Top level Reconcile logic:

- If generation number hasn't changed, do nothing. We don't want to process updates
  unless spec has changed.

- Add a finalizer if not already present. This will convert deletes to updates
  and allows us to perform any actions before the resource is actually deleted.
  However, there is really no delete logic at present.

- Check if the annotation, which indicates that repo is already initialized, is
  present. If so, there is nothing more to do. If not, proceed with init logic.

- Since we don't generate a unique name for init pod, it is possible that pod
  from a previous attempt still exists. So check for such a pod and delete it.
  We may eventually use unique names but that requires clean up of old pods.
  Also note that the name of the pod includes BackupLocation resource name so it
  is not a hard-coded name really.

- Create the pod that will initialize the repo. The kubedrutil "repoinit" command
  will call restic to initialize the repo and it will then set the annotation to
  indicate that repo is initialized. It will also set the status both in case of
  success and failure.
*/

func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// In case of some errors such as "not found" and "already exists",
// there is no point in requeuing the reconcile.
// See https://github.com/kubernetes-sigs/controller-runtime/issues/377
func ignoreErrors(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}

	if apierrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (r *BackupLocationReconciler) setStatus(backupLoc *kubedrv1alpha1.BackupLocation, status string, errmsg string) {
	// Allows us to check and skip reconciles for only metadata updates.
	backupLoc.Status.ObservedGeneration = backupLoc.ObjectMeta.Generation

	backupLoc.Status.InitStatus = status
	if errmsg == "" {
		// For some reason, empty error string is causing problems even though
		// the field is marked "optional" in Status struct.
		errmsg = "."
	}
	backupLoc.Status.InitErrorMessage = errmsg

	backupLoc.Status.InitTime = metav1.Now().String()

	r.Log.Info("Updating status...")
	if err := r.Status().Update(context.Background(), backupLoc); err != nil {
		r.Log.Error(err, "unable to update backup location status")
	}
}

// Reconcile is the the main entry point called by the framework.
func (r *BackupLocationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("backuplocation", req.NamespacedName)

	var backupLoc kubedrv1alpha1.BackupLocation
	if err := r.Get(ctx, req.NamespacedName, &backupLoc); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification).
			log.Info("BackupLocation (" + req.NamespacedName.Name + ") is not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch BackupLocation")
		return ctrl.Result{}, err
	}

	// Skip if spec hasn't changed. This check prevents reconcile on status
	// updates.
	if backupLoc.Status.ObservedGeneration == backupLoc.ObjectMeta.Generation {
		r.Log.Info("Skipping reconcile as generation number hasn't changed")
		return ctrl.Result{}, nil
	}

	finalizer := "backuplocation.finalizers.kubedr.catalogicsoftware.com"

	if backupLoc.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !containsString(backupLoc.ObjectMeta.Finalizers, finalizer) {
			backupLoc.ObjectMeta.Finalizers = append(backupLoc.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), &backupLoc); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(backupLoc.ObjectMeta.Finalizers, finalizer) {
			// our finalizer is present, handle any pre-deletion logic here.

			// remove our finalizer from the list and update it.
			backupLoc.ObjectMeta.Finalizers = removeString(backupLoc.ObjectMeta.Finalizers, finalizer)

			if err := r.Update(context.Background(), &backupLoc); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Nothing more to do for DELETE.
		return ctrl.Result{}, nil
	}

	// Check annotations to see if repo is already initialized.
	// Ideally, we should check the repo itself to confirm that it is
	// initialized, instead of depending on annotation.
	initAnnotation := "initialized.annotations.kubedr.catalogicsoftware.com"

	initialized, exists := backupLoc.ObjectMeta.Annotations[initAnnotation]
	if exists && (initialized == "true") {
		// No need to initialize the repo.
		log.Info("Repo is already initialized")
		return ctrl.Result{}, nil
	}

	// Annotation doesn't exist so we need to initialize the repo.

	initPodName := backupLoc.Name + "-init-pod"

	// Since we don't generate a unique name for the pod that initializes the repo,
	// we need to explicitly check and delete the pod if it exists. We may eventually
	// use a unique name but that will also require cleanup of old pods.
	var pod corev1.Pod
	if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: initPodName}, &pod); err == nil {
		log.Info("Found init pod, will delete it and continue...")
		if err := r.Delete(ctx, &pod); ignoreNotFound(err) != nil {
			log.Error(err, "Error in deleting init pod")
			return ctrl.Result{}, err
		}
	}

	r.setStatus(&backupLoc, "Initializing", "")

	// Initialize the repo.
	initPod, err := buildResticRepoInitPod(&backupLoc, log)
	if err != nil {
		log.Error(err, "Error in creating init pod")
		return ctrl.Result{}, err
	}

	if err := ctrl.SetControllerReference(&backupLoc, initPod, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Creating a new Pod", "Pod.Namespace", initPod.Namespace, "Pod.Name", initPod.Name)
	err = r.Create(ctx, initPod)
	if err != nil {
		r.setStatus(&backupLoc, "Failed", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func buildResticRepoInitPod(cr *kubedrv1alpha1.BackupLocation, log logr.Logger) (*corev1.Pod, error) {
	kubedrUtilImage := os.Getenv("KUBEDR_UTIL_IMAGE")
	if kubedrUtilImage == "" {
		// This should really not happen.
		err := fmt.Errorf("KUBEDR_UTIL_IMAGE is not set")
		log.Error(err, "")
		return nil, err
	}
	log.V(1).Info(fmt.Sprintf("kubedrUtilImage: %s", kubedrUtilImage))

	s3EndPoint := "s3:" + cr.Spec.Url + "/" + cr.Spec.BucketName

	labels := map[string]string{
		"kubedr.type":      "backuploc-init",
		"kubedr.backuploc": cr.Name,
	}

	accessKey := corev1.SecretKeySelector{}
	accessKey.Name = cr.Spec.Credentials
	accessKey.Key = "access_key"

	secretKey := corev1.SecretKeySelector{}
	secretKey.Name = cr.Spec.Credentials
	secretKey.Key = "secret_key"

	resticPassword := corev1.SecretKeySelector{}
	resticPassword.Name = cr.Spec.Credentials
	resticPassword.Key = "restic_repo_password"

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-init-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  cr.Name + "-init",
					Image: kubedrUtilImage,
					Args: []string{
						"/usr/local/bin/kubedrutil", "repoinit",
					},
					Env: []corev1.EnvVar{
						{
							Name: "MY_POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "AWS_ACCESS_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &accessKey,
							},
						},
						{
							Name: "AWS_SECRET_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &secretKey,
							},
						},
						{
							Name: "RESTIC_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &resticPassword,
							},
						},
						{
							Name:  "RESTIC_REPO",
							Value: s3EndPoint,
						},
						{
							Name:  "KDR_BACKUPLOC_NAME",
							Value: cr.Name,
						},
					},
				},
			},
			RestartPolicy: "Never",
		},
	}, nil
}

// SetupWithManager hooks up this controller with the manager.
func (r *BackupLocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubedrv1alpha1.BackupLocation{}).
		Complete(r)
}
