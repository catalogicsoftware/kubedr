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

	kubedrv1alpha1 "kubedr/api/v1alpha1"
)

// MetadataRestoreReconciler reconciles a MetadataRestore object
type MetadataRestoreReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *MetadataRestoreReconciler) setStatus(mr *kubedrv1alpha1.MetadataRestore, status string, errmsg string) {
	mr.Status.ObservedGeneration = mr.ObjectMeta.Generation

	mr.Status.RestoreStatus = status
	mr.Status.RestoreErrorMessage = errmsg
	mr.Status.RestoreTime = metav1.Now().String()

	r.Log.Info("Updating status...")
	if err := r.Status().Update(context.Background(), mr); err != nil {
		r.Log.Error(err, "unable to update MetadataRestore status")
	}
}

// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatarestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatarestores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuprecords/status,verbs=get
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=backuplocations/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=pods,verbs=create;get

/*
 * Top level Reconcile logic
 *
 * - If generation number hasn't changed, do nothing. We don't want to process updates
 *   unless spec has changed.
 *
 * - Check if the annotation, which indicates that this restore resource is already
 *   processed, is present. If so, there is nothing more to do. If not, proceed with
 *   restore logic.
 *
 * - There is nothing to do for deletion so we don't add any finalizers.
 *
 * - If there is a previous restore pod for this resource, delete the pod.
 *
 * - Create the pod that will restore the data. The kubedrutil "restore" command
 *   will call restic to restore the data and then, it will set the annotation to
 *   indicate that this resource is processed.
 *
 * - The "restore" command will also set the status both in case of success and
 *   failure.
 */

// Reconcile is the the main entry point called by the framework.
func (r *MetadataRestoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var mr kubedrv1alpha1.MetadataRestore
	if err := r.Get(ctx, req.NamespacedName, &mr); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification).
			r.Log.Info("MetadataRestore (" + req.NamespacedName.Name + ") is not found")
			return ctrl.Result{}, nil
		}

		r.Log.Error(err, "unable to fetch MetadataRestore")
		return ctrl.Result{}, err
	}

	// Skip if spec hasn't changed. This check prevents reconcile on status
	// updates.
	if mr.Status.ObservedGeneration == mr.ObjectMeta.Generation {
		r.Log.Info("Skipping reconcile as generation number hasn't changed")
		return ctrl.Result{}, nil
	}

	// No deletion logic as we don't really have anything to do during
	// deletion of a MetadataRestore resource.

	// Check annotations to see if this resource was already processed
	// and restore was successful.
	restoreAnnotation := "restored.annotations.kubedr.catalogicsoftware.com"

	restored, exists := mr.ObjectMeta.Annotations[restoreAnnotation]
	if exists && (restored == "true") {
		// No need to process the resource as restore was done already.
		r.Log.Info("Restore was already done")
		return ctrl.Result{}, nil
	}

	// We are deliberately avoiding any attempt to make the name unique.
	// The client is in a better position to come up with a unique name.
	// If we do switch to generating a unique name, we need to make sure
	// that any previous pods are cleaned up.
	podName := mr.Name + "-mr"

	// Since we don't generate a unique name for the pod that initializes the repo,
	// we need to explicitly check and delete the pod if it exists.
	var prevPod corev1.Pod
	if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: podName}, &prevPod); err == nil {
		r.Log.Info("Found a previous restore pod, will delete it and continue...")
		if err := r.Delete(ctx, &prevPod); ignoreNotFound(err) != nil {
			r.Log.Error(err, "Error in deleting init pod")
			return ctrl.Result{}, err
		}
	}

	pod, err := r.buildRestorePod(&mr, req.Namespace, podName)
	if err != nil {
		r.Log.Error(err, "Error in creating restore pod")
		if apierrors.IsNotFound(err) {
			// This shouldn't really happen but if an invalid MBR is given or
			// if backup location inside the MBR is wrong, there is nothing we can
			// do.
			r.setStatus(&mr, "Failed", "Error in creating restore pod")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := ctrl.SetControllerReference(&mr, pod, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Starting a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	err = r.Create(ctx, pod)
	if err != nil {
		r.Log.Error(err, "Error in starting restore pod")
		r.setStatus(&mr, "Failed", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager hooks up this controller with the manager.
func (r *MetadataRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubedrv1alpha1.MetadataRestore{}).
		Complete(r)
}

func getRepoData(backupLocation *kubedrv1alpha1.BackupLocation) (string, *corev1.SecretKeySelector,
	*corev1.SecretKeySelector, *corev1.SecretKeySelector) {

	s3EndPoint := "s3:" + backupLocation.Spec.Url + "/" + backupLocation.Spec.BucketName

	accessKey := corev1.SecretKeySelector{}
	accessKey.Name = backupLocation.Spec.Credentials
	accessKey.Key = "access_key"

	secretKey := corev1.SecretKeySelector{}
	secretKey.Name = backupLocation.Spec.Credentials
	secretKey.Key = "secret_key"

	resticPassword := corev1.SecretKeySelector{}
	resticPassword.Name = backupLocation.Spec.Credentials
	resticPassword.Key = "restic_repo_password"

	return s3EndPoint, &accessKey, &secretKey, &resticPassword
}

func (r *MetadataRestoreReconciler) buildRestorePod(cr *kubedrv1alpha1.MetadataRestore,
	namespace string, podName string) (*corev1.Pod, error) {

	kubedrUtilImage := os.Getenv("KUBEDR_UTIL_IMAGE")
	if kubedrUtilImage == "" {
		// This should really not happen.
		err := fmt.Errorf("KUBEDR_UTIL_IMAGE is not set")
		r.Log.Error(err, "")
		return nil, err
	}

	mbr := &kubedrv1alpha1.MetadataBackupRecord{}
	mbrKey := types.NamespacedName{Namespace: namespace, Name: cr.Spec.MBRName}
	if err := r.Get(context.TODO(), mbrKey, mbr); err != nil {
		return nil, err
	}

	backupLocation := &kubedrv1alpha1.BackupLocation{}
	backupLocKey := types.NamespacedName{Namespace: namespace, Name: mbr.Spec.Backuploc}
	if err := r.Get(context.TODO(), backupLocKey, backupLocation); err != nil {
		return nil, err
	}
	s3EndPoint, accessKey, secretKey, resticPassword := getRepoData(backupLocation)

	labels := map[string]string{
		"kubedr.type":        "restore",
		"kubedr.restore-mbr": mbr.Name,
	}

	targetDirVolume := corev1.Volume{Name: "restore-target"}
	targetDirVolume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: cr.Spec.PVCName}

	volumes := []corev1.Volume{
		targetDirVolume,
	}

	env := []corev1.EnvVar{
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
				SecretKeyRef: accessKey,
			},
		},
		{
			Name: "AWS_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: secretKey,
			},
		},
		{
			Name: "RESTIC_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: resticPassword,
			},
		},
		{
			Name:  "KDR_MR_NAME",
			Value: cr.Name,
		},
		{
			Name:  "RESTIC_REPO",
			Value: s3EndPoint,
		},
		{
			Name:  "KDR_RESTORE_DEST",
			Value: "/restore",
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "restore-target",
			MountPath: "/restore",
		},
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: corev1.PodSpec{
			RestartPolicy: "Never",

			Volumes: volumes,

			Containers: []corev1.Container{
				{
					Name:         cr.Name,
					Image:        kubedrUtilImage,
					VolumeMounts: volumeMounts,
					Env:          env,

					Args: []string{
						"/usr/local/bin/kubedrutil", "restore",
					},
				},
			},
		},
	}, nil
}
