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
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubedrv1alpha1 "kubedr/api/v1alpha1"
)

// MetadataBackupPolicyReconciler reconciles a MetadataBackupPolicy object
type MetadataBackupPolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Implements logic to handle a new policy.
func (r *MetadataBackupPolicyReconciler) newPolicy(policy *kubedrv1alpha1.MetadataBackupPolicy,
	namespace string, cronJobName string) (ctrl.Result, error) {

	r.setStatus(policy)

	backupCronjob, err := r.buildBackupCronjob(policy, namespace, cronJobName)
	if err != nil {
		r.Log.Error(err, "Error in creating backup cronjob")
		return ctrl.Result{}, err
	}

	// Set Policy as owner of cronjob so that when policy is deleted,
	// cronjob is cleaned up automatically.
	if err := ctrl.SetControllerReference(policy, backupCronjob, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Creating a new Cronjob", "Namespace", backupCronjob.Namespace, "Name", backupCronjob.Name)
	err = r.Create(context.Background(), backupCronjob)
	if err != nil {
		r.Log.Info(err.Error())
		return ctrl.Result{}, ignoreErrors(err)
	}

	if err == nil {
		// We have seen a second reconcile request immediately after return from
		// here and add to it the fact that Get() is failing with "not found" errors
		// even though the resource has just been created (Get reads from local cache).
		// So make sure cache is updated before returning from here.
		//
		// We are just waiting for some time. Does it ensure that cache is updated?
		// Need to know more about cache semantics.
		r.waitForCreatedResource(namespace, cronJobName)
	}

	return ctrl.Result{}, err
}

// Policy and Cronjob already exist. Make any required changes to the cronjob.
// If retention is changed, there is nothing to be done here. The retention
// logic in in MetadataBackupRecord controller.
func (r *MetadataBackupPolicyReconciler) processUpdate(policy *kubedrv1alpha1.MetadataBackupPolicy,
	cronJob *batchv1beta1.CronJob) (ctrl.Result, error) {

	updateCron := false

	if cronJob.Spec.Schedule != policy.Spec.Schedule {
		r.Log.Info("Schedule changed")
		cronJob.Spec.Schedule = policy.Spec.Schedule
		updateCron = true
	}

	if cronJob.Spec.Suspend != policy.Spec.Suspend {
		r.Log.V(1).Info("suspend status changed")
		cronJob.Spec.Suspend = policy.Spec.Suspend
		updateCron = true
	}

	if updateCron {
		if err := r.Update(context.Background(), cronJob); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// Process spec and make sure it matches status of the world.
func (r *MetadataBackupPolicyReconciler) processSpec(policy *kubedrv1alpha1.MetadataBackupPolicy,
	namespace string) (ctrl.Result, error) {

	var cronJob batchv1beta1.CronJob
	cronJobName := policy.Name + "-backup-cronjob"

	// I have seen Get return "not found" and then the following
	// create fail with "already exists" error.
	if err := r.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: cronJobName},
		&cronJob); err != nil {

		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		return r.newPolicy(policy, namespace, cronJobName)
	}

	// The policy exists. We need to check and make any required changes to cronJob.
	return r.processUpdate(policy, &cronJob)
}

func (r *MetadataBackupPolicyReconciler) setStatus(policy *kubedrv1alpha1.MetadataBackupPolicy) {
	policy.Status.BackupStatus = "Initializing"
	policy.Status.BackupTime = metav1.Now().String()

	r.Log.Info("Updating status...")
	if err := r.Status().Update(context.Background(), policy); err != nil {
		r.Log.Error(err, "unable to update policy status")
	}
}

// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=create;get;list;update;patch;delete;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=create;get;list;update;patch;delete;watch

// Reconcile is the the main entry point called by the framework.
func (r *MetadataBackupPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("metadatabackuppolicy", req.NamespacedName)
	r.Log = log

	var policy kubedrv1alpha1.MetadataBackupPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification).
			log.Info("MetadataBackupPolicy (" + req.NamespacedName.Name + ") is not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch MetadataBackupPolicy")
		return ctrl.Result{}, err
	}

	finalizer := "metadata-backup-policy.finalizers.kubedr.catalogicsoftware.com"

	if policy.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !containsString(policy.ObjectMeta.Finalizers, finalizer) {
			policy.ObjectMeta.Finalizers = append(policy.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), &policy); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(policy.ObjectMeta.Finalizers, finalizer) {
			// our finalizer is present, handle any pre-deletion logic here.

			// remove our finalizer from the list and update it.
			policy.ObjectMeta.Finalizers = removeString(policy.ObjectMeta.Finalizers, finalizer)

			if err := r.Update(context.Background(), &policy); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Nothing more to do for DELETE.
		return ctrl.Result{}, nil
	}

	return r.processSpec(&policy, req.Namespace)
}

// SetupWithManager hooks up this controller with the manager.
func (r *MetadataBackupPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubedrv1alpha1.MetadataBackupPolicy{}).
		Owns(&batchv1beta1.CronJob{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *MetadataBackupPolicyReconciler) waitForCreatedResource(namespace string, name string) {
	var cronJob batchv1beta1.CronJob

	for i := 0; i < 5; i++ {
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &cronJob)
		if err == nil {
			return
		}

		time.Sleep(2 * time.Second)
	}
}

func (r *MetadataBackupPolicyReconciler) getMasterNodeLabelName(policy *kubedrv1alpha1.MetadataBackupPolicy) string {
	labelName := "node-role.kubernetes.io/master"

	if policy.Spec.Options == nil {
		return labelName
	}

	key := "master-node-label-name"

	val, exists := policy.Spec.Options[key]
	if !exists {
		return labelName
	}

	if len(val) == 0 {
		r.Log.Error(fmt.Errorf("Invalid value for master node label name in config map (%s)",
			policy.Spec.Options), "")
		return labelName
	}

	r.Log.Info(fmt.Sprintf("master node label name in config map (%s): %s", policy.Spec.Options, val))

	return val
}

func (r *MetadataBackupPolicyReconciler) buildBackupCronjob(cr *kubedrv1alpha1.MetadataBackupPolicy,
	namespace string, cronJobName string) (*batchv1beta1.CronJob, error) {

	labels := map[string]string{
		"kubedr.type":          "backup",
		"kubedr.backup-policy": cr.Name,
	}

	kubedrUtilImage := os.Getenv("KUBEDR_UTIL_IMAGE")
	if kubedrUtilImage == "" {
		// This should really not happen.
		err := fmt.Errorf("KUBEDR_UTIL_IMAGE is not set")
		r.Log.Error(err, "")
		return nil, err
	}
	r.Log.V(1).Info(fmt.Sprintf("kubedrUtilImage: %s", kubedrUtilImage))

	backupLocation := &kubedrv1alpha1.BackupLocation{}
	backupLocKey := types.NamespacedName{Namespace: namespace, Name: cr.Spec.Destination}
	err := r.Get(context.TODO(), backupLocKey, backupLocation)
	if err != nil {
		// If the error is "not found", there is no point in retrying.
		return nil, err
	}

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

	targetDirVolume := corev1.Volume{Name: "target-dir"}
	targetDirVolume.EmptyDir = &corev1.EmptyDirVolumeSource{}

	etcdCredsVolume := corev1.Volume{Name: "etcd-creds"}
	etcdCredsVolume.Secret = &corev1.SecretVolumeSource{
		SecretName: cr.Spec.EtcdCreds,
	}

	volumes := []corev1.Volume{
		targetDirVolume,
		etcdCredsVolume,
	}

	env := []corev1.EnvVar{
		{
			Name: "MY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource {
				FieldRef: &corev1.ObjectFieldSelector {
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
			Name:  "KDR_POLICY_NAME",
			Value: cr.Name,
		},
		{
			Name:  "ETCD_ENDPOINT",
			Value: cr.Spec.EtcdEndpoint,
		},
		{
			Name:  "ETCD_CREDS_DIR",
			Value: "/etcd_creds",
		},
		{
			Name:  "ETCD_SNAP_PATH",
			Value: "/data/etcd-snapshot.db",
		},
		{
			Name:  "RESTIC_REPO",
			Value: s3EndPoint,
		},
		{
			Name:  "BACKUP_SRC",
			Value: "/data",
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "target-dir",
			MountPath: "/data",
		},
		{
			Name:      "etcd-creds",
			MountPath: "/etcd_creds",
		},
	}

	// Certs dir is optional and if not given, do not pass details to the
	// backup pod/container.
	if cr.Spec.CertsDir != "" {
		var t corev1.HostPathType = "Directory"
		certsDirVolume := corev1.Volume{Name: "certs-dir"}
		certsDirVolume.HostPath = &corev1.HostPathVolumeSource{
			Path: cr.Spec.CertsDir,
			Type: &t,
		}

		volumes = append(volumes, certsDirVolume)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "certs-dir", MountPath: "/certs_dir"})

		// FIX: Shouldn't hard code /data here. It ties too closely with the
		// code in the backup container.
		env = append(env, corev1.EnvVar{Name: "CERTS_DEST_DIR", Value: "/data/certificates"})
		env = append(env, corev1.EnvVar{Name: "CERTS_SRC_DIR", Value: "/certs_dir"})
	}

	masterNodeLabelName := r.getMasterNodeLabelName(cr)

	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},

		Spec: batchv1beta1.CronJobSpec{
			ConcurrencyPolicy: "Forbid",
			Schedule:          cr.Spec.Schedule,
			Suspend:           cr.Spec.Suspend,

			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name + "-backup-job",
					Namespace: cr.Namespace,
				},
				Spec: batchv1.JobSpec{
					// TODO: Set backoffLimit to 2.
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      cr.Name + "-backup-pod-template",
							Namespace: cr.Namespace,
							Labels:    labels,
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							HostNetwork:   true,

							// Make sure that backup pod runs on the master.
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											corev1.NodeSelectorTerm{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													corev1.NodeSelectorRequirement{
														Key:      masterNodeLabelName,
														Operator: "Exists",
													},
												},
											},
										},
									},
								},
							},

							// Tolerate "NoSchedule" taint on master nodes.
							Tolerations: []corev1.Toleration{
								corev1.Toleration{
									Operator: "Exists",
									Effect:   "NoSchedule",
								},
							},

							Volumes: volumes,

							Containers: []corev1.Container{
								{
									Name:         cr.Name + "-kcx-backup",
									Image:        kubedrUtilImage,
									VolumeMounts: volumeMounts,
									Env:          env,

									Args: []string{
										"/usr/local/bin/kubedrutil", "backup",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
