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

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.
*/
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// In case of some errors such as "not found" and "already exist",
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

func (r *BackupLocationReconciler) setStatus(backupLoc *kubedrv1alpha1.BackupLocation, status string, errmsg string) error {
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
		return err
	}

	return nil
}

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
	init_annotation := "initialized.annotations.kubedr.catalogicsoftware.com"

	initialized, exists := backupLoc.ObjectMeta.Annotations[init_annotation]
	if exists && (initialized == "true") {
		// No need to initialize the repo.
		log.Info("Repo is already initialized")
		return ctrl.Result{}, nil
	}

	// Check if init pod is already created and if so, return.
	var pod corev1.Pod
	if err := r.Get(ctx,
		types.NamespacedName{Namespace: req.Namespace, Name: backupLoc.Name + "-init-pod"},
		&pod); err == nil {

		return ctrl.Result{}, nil
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
		// Do nothing if pod creation failed because there is already an existing
		// init pod. This check prevents cycles.
		if apierrors.IsAlreadyExists(err) {
			log.Info("Creation of init pod failed because it already exists")
			r.setStatus(&backupLoc, "Failed", "Init pod already exists")
		} else {
			r.setStatus(&backupLoc, "Failed", err.Error())
			return ctrl.Result{}, err
		}
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
							Name: "AWS_ACCESS_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &access_key,
							},
						},
						{
							Name: "AWS_SECRET_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &secret_key,
							},
						},
						{
							Name: "RESTIC_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &restic_password,
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

func (r *BackupLocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubedrv1alpha1.BackupLocation{}).
		Complete(r)
}
