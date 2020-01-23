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
	"k8s.io/apimachinery/pkg/types"
	"sort"

	//	batchv1 "k8s.io/api/batch/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/prometheus/client_golang/prometheus"
	kubedrv1alpha1 "kubedr/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	numDeletedBackups = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kubedr_num_metadata_backups_deleted",
			Help: "Number of metadata backups deleted",
		},
	)
)

// MetadataBackupRecordReconciler reconciles a MetadataBackupRecord object
type MetadataBackupRecordReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuprecords,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuprecords/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=create;get;list;update;patch;delete;watch

func (r *MetadataBackupRecordReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("metadatabackuprecord", req.NamespacedName)

	// Every time a MBR is created, we need to check and delete some older snapshot
	// as per retention setting.

	var record kubedrv1alpha1.MetadataBackupRecord
	if err := r.Get(ctx, req.NamespacedName, &record); err != nil {
		log.Error(err, "unable to fetch MetadataBackupRecord")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification).
		return ctrl.Result{}, ignoreNotFound(err)
	}

	finalizer := "mbr.finalizers.kubedr.catalogicsoftware.com"

	if record.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !containsString(record.ObjectMeta.Finalizers, finalizer) {
			record.ObjectMeta.Finalizers = append(record.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), &record); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(record.ObjectMeta.Finalizers, finalizer) {
			// Our finalizer is present, handle any pre-deletion logic here.

			// remove our finalizer from the list and update it.
			record.ObjectMeta.Finalizers = removeString(record.ObjectMeta.Finalizers, finalizer)

			if err := r.Update(context.Background(), &record); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Nothing more to do for DELETE.
		return ctrl.Result{}, nil
	}

	var policy kubedrv1alpha1.MetadataBackupPolicy
	log.Info("Getting policy...")
	if err := r.Get(ctx,
		types.NamespacedName{Namespace: req.Namespace, Name: record.Spec.Policy},
		&policy); err != nil {

		log.Error(err, "unable to fetch MetadataBackupPolicy, no retention processing")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification).
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// Now, make sure spec matches the status of world.
	log.Info("Getting MBR list...")
	var mbrList kubedrv1alpha1.MetadataBackupRecordList
	if err := r.List(ctx, &mbrList, client.InNamespace(req.Namespace),
		client.MatchingFields{"policy": record.Spec.Policy}); err != nil {

		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Number of MBR entries: %d", len(mbrList.Items)))

	sort.Slice(mbrList.Items, func(i, j int) bool {
		return mbrList.Items[i].ObjectMeta.CreationTimestamp.Before(&mbrList.Items[j].ObjectMeta.CreationTimestamp)
	})

	log.Info(fmt.Sprintf("retention: %d", *policy.Spec.RetainNumBackups))

	if int64(len(mbrList.Items)) <= *policy.Spec.RetainNumBackups {
		log.Info("Number of backups is less than retention...")
		return ctrl.Result{}, nil
	}

	backupLoc := &kubedrv1alpha1.BackupLocation{}
	backupLocKey := types.NamespacedName{Namespace: req.Namespace, Name: policy.Spec.Destination}
	err := r.Get(context.TODO(), backupLocKey, backupLoc)
	if err != nil {
		// If the error is "not found", there is no point in retrying.
		return ctrl.Result{}, err
	}

	// There are some snapshots that need to be deleted.
	for i := 0; int64(i) < (int64(len(mbrList.Items)) - *policy.Spec.RetainNumBackups); i++ {
		log.Info("Need to delete: " + mbrList.Items[i].Spec.SnapshotId)

		// Delete the record first.
		if err := r.Delete(ctx, &mbrList.Items[i]); ignoreNotFound(err) != nil {
			log.Error(err, "unable to delete mbr", "mbr", mbrList.Items[i])
		} else {
			log.V(0).Info("deleted mbr", "mbr", mbrList.Items[i])
		}

		pod, err := createResticSnapDeletePod(backupLoc, log, mbrList.Items[i].Spec.SnapshotId,
			mbrList.Items[i].Name, mbrList.Items[i].Namespace)

		if err != nil {
			log.Error(err, "Error in creating snapshot deletion pod")
			return ctrl.Result{}, err
		}

		// Delete restic snapshot
		log.Info("Starting a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Error in starting snap delete pod")
			return ctrl.Result{}, err
		}

		// FIX: We really need to make sure that delete succeeded.
		numDeletedBackups.Inc()
	}

	// Keep last 3 snap deletetion pods and clean up the rest.
	// Make this number configurable. We need global options. This is not related
	// to individual policies.
	r.cleanupOldSnapDeletionPods(req.Namespace, log)

	return ctrl.Result{}, nil
}

func (r *MetadataBackupRecordReconciler) cleanupOldSnapDeletionPods(namespace string, log logr.Logger) {
	ctx := context.Background()

	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(namespace),
		client.MatchingLabels{"kubedr.catalogicsoftware.com/snap-deletion-pod": "true"}); err != nil {
		log.Error(err, "unable to list snap deletion pods")
		return
	}

	log.Info(fmt.Sprintf("Number of snap deletion pods: %d", len(podList.Items)))

	sort.Slice(podList.Items, func(i, j int) bool {
		return podList.Items[i].ObjectMeta.CreationTimestamp.Before(&podList.Items[j].ObjectMeta.CreationTimestamp)
	})

	if int64(len(podList.Items)) <= 3 {
		return
	}

	// There are some pods that need to be deleted.
	for i := 0; int64(i) < (int64(len(podList.Items)) - 3); i++ {
		if err := r.Delete(ctx, &podList.Items[i]); ignoreNotFound(err) != nil {
			log.Error(err, "unable to delete pod", "pod", podList.Items[i].Name)
		} else {
			log.V(0).Info("deleted pod", "pod", podList.Items[i].Name)
		}
	}
}

func (r *MetadataBackupRecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&kubedrv1alpha1.MetadataBackupRecord{},
		"policy", func(rawObj runtime.Object) []string {
			// grab the job object, extract the owner...
			record := rawObj.(*kubedrv1alpha1.MetadataBackupRecord)

			return []string{record.Spec.Policy}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kubedrv1alpha1.MetadataBackupRecord{}).
		Complete(r)
}

func createResticSnapDeletePod(backupLocation *kubedrv1alpha1.BackupLocation, log logr.Logger,
	snapshotId string, mbrName string, namespace string) (*corev1.Pod, error) {

	s3EndPoint := "s3:" + backupLocation.Spec.Url + "/" + backupLocation.Spec.BucketName

	access_key := corev1.SecretKeySelector{}
	access_key.Name = backupLocation.Spec.Credentials
	access_key.Key = "access_key"

	secret_key := corev1.SecretKeySelector{}
	secret_key.Name = backupLocation.Spec.Credentials
	secret_key.Key = "secret_key"

	restic_password := corev1.SecretKeySelector{}
	restic_password.Name = backupLocation.Spec.Credentials
	restic_password.Key = "restic_repo_password"

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mbrName + "-snapdel-pod-" + snapshotId,
			Namespace: namespace,
			Labels: map[string]string{
				"kubedr.catalogicsoftware.com/snap-deletion-pod": "true",
			},
		},

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  mbrName + "-del",
					Image: "restic/restic",
					Args:  []string{"-r", s3EndPoint, "forget", "--prune", snapshotId},
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
					},
				},
			},
			RestartPolicy: "Never",
		},
	}, nil
}

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(numDeletedBackups)
}
