/*

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
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=create;get;list;update;patch;delete;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=create;get;list;update;patch;delete;watch

func (r *MetadataBackupPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("metadatabackuppolicy", req.NamespacedName)

	var policy kubedrv1alpha1.MetadataBackupPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		log.Error(err, "unable to fetch MetadataBackupPolicy")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification).
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// RD: Make the exact value configurable.
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
			// TODO

            // remove our finalizer from the list and update it.
            policy.ObjectMeta.Finalizers = removeString(policy.ObjectMeta.Finalizers, finalizer)
 
			if err := r.Update(context.Background(), &policy); err != nil {
                return ctrl.Result{}, err
            }
        }

		// Nothing more to do for DELETE.
        return ctrl.Result{}, nil
    }

	// TODO: Update Status.

	// Now, make sure spec matches the status of world.

	var cronJob batchv1beta1.CronJob
	cronJobName := policy.Name + "-backup-cronjob"

	// I have seen Get return "not found" and then the following
	// create fail with "already exists" error.
	
	if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: cronJobName}, &cronJob); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
			
		// Cronjob doesn't exist, create one and return.
		backupCronjob, err := r.buildBackupCronjob(&policy, req.Namespace, cronJobName, log)
		if err != nil {
			log.Error(err, "Error in creating backup cronjob")
			return ctrl.Result{}, err
		}

		// Set Policy as owner of cronjob so that when policy is deleted,
		// cronjob is cleaned up automatically.
		if err := ctrl.SetControllerReference(&policy, backupCronjob, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Cronjob", "Namespace", backupCronjob.Namespace, "Name", backupCronjob.Name)
		err = r.Create(ctx, backupCronjob)
		if err != nil {
			log.Info(err.Error())
			return ctrl.Result{}, ignoreErrors(err)
		}

		if err == nil {
			// We have seen a second reconcile request immediately after return from 
			// here and add to it the fact that Get() is failing with "not found" errors
			// even though the resource has just been created (Get reads from local cache).
			// So make sure cache is updated before returning from here.
			//
			// We are just waiting for some time. Does it ensure that cache is updated?
			// TODO: Need to know more about cache semantics.
			r.waitForCreatedResource(req.Namespace, cronJobName)
		}

		return ctrl.Result{}, err
	}

	// cron job exists, check if we need to make any changes to its spec.
	// For now, only support update of schedule.

	updateCron := false

	if cronJob.Spec.Schedule != policy.Spec.Schedule {
		log.Info("Schedule changed")
		cronJob.Spec.Schedule = policy.Spec.Schedule
		updateCron = true
	}

	if cronJob.Spec.Suspend != policy.Spec.Suspend {
		log.V(1).Info("suspend status changed")
		cronJob.Spec.Suspend = policy.Spec.Suspend
		updateCron = true
	}

	if updateCron {
		if err := r.Update(ctx, &cronJob); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

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

func (r *MetadataBackupPolicyReconciler) getMasterNodeLabelName(policy *kubedrv1alpha1.MetadataBackupPolicy, 
	log logr.Logger) string {

	labelName := "node-role.kubernetes.io/master"

	if policy.Spec.Options == nil {
		return labelName;
	}

/*
	options := &corev1.ConfigMap{}
	optionsKey := types.NamespacedName{Namespace: policy.Namespace, Name: policy.Spec.Options}

	err := r.Get(context.TODO(), optionsKey, options)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Error in getting config map (%s)", policy.Spec.Options))
		}

		return labelName;
	}

	if options.Data == nil {
		log.Error(err, fmt.Sprintf("No Data in config map (%s)", policy.Spec.Options))
		return labelName;
	}
*/

	key := "master-node-label-name"

	// val, exists := options.Data[key]
	val, exists := policy.Spec.Options[key]
	if !exists {
		return labelName;
	}

	if len(val) == 0 {
		log.Error(errors.New(fmt.Sprintf("Invalid value for master node label name in config map (%s)", policy.Spec.Options)),
		"")
		return labelName;
	}

	log.Info(fmt.Sprintf("master node label name in config map (%s): %s", policy.Spec.Options, val))

	return val
}

func (r *MetadataBackupPolicyReconciler) buildBackupCronjob(cr  *kubedrv1alpha1.MetadataBackupPolicy, 
	namespace string, cronJobName string, log logr.Logger) (*batchv1beta1.CronJob, error) {

	backupLocation := &kubedrv1alpha1.BackupLocation{}
	backupLocKey := types.NamespacedName{Namespace: namespace, Name: cr.Spec.Destination}
	err := r.Get(context.TODO(), backupLocKey, backupLocation)
	if err != nil {
		// RD: If the error is "not found", there is no point in retrying.
		return nil, err
	}

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

	targetDirVolume := corev1.Volume{Name: "target-dir"}
	targetDirVolume.EmptyDir = &corev1.EmptyDirVolumeSource {}

	etcdCredsVolume := corev1.Volume{Name: "etcd-creds"}
	etcdCredsVolume.Secret = &corev1.SecretVolumeSource {
		SecretName: cr.Spec.EtcdCreds,
	}

	volumes := [] corev1.Volume {
		targetDirVolume,
		etcdCredsVolume,
	}

	// This should ideally be done in defaulter web hook.
	env := []corev1.EnvVar {
		{
			Name: "AWS_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource {
				SecretKeyRef: &access_key,
			},
		},
		{
			Name: "AWS_SECRET_KEY",
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
		{
			Name: "KDR_POLICY_NAME",
			Value: cr.Name,
		},
		{
			Name: "ETCD_ENDPOINT",
			Value: cr.Spec.EtcdEndpoint,
		},
		{
			Name: "ETCD_CREDS_DIR",
			Value: "/etcd_creds",
		},
		{
			Name: "ETCD_SNAP_PATH",
			Value: "/data/etcd-snapshot.db",
		},
		{
			Name: "RESTIC_REPO",
			Value: s3EndPoint,
		},
		{
			Name: "BACKUP_SRC",
			Value: "/data",
		},
	}

	volumeMounts := []corev1.VolumeMount {
		{
			Name: "target-dir",
			MountPath: "/data",
		},
		{
			Name: "etcd-creds",
			MountPath: "/etcd_creds",
		},
	}

	// Certs dir is optional and if not given, do not pass details to the
	// backup pod/container.
	if cr.Spec.CertsDir != "" {
		var t corev1.HostPathType = "Directory"
		certsDirVolume := corev1.Volume{Name: "certs-dir"}
		certsDirVolume.HostPath = &corev1.HostPathVolumeSource {
			Path: cr.Spec.CertsDir,
			Type: &t,
		}
		
		volumes = append(volumes, certsDirVolume)
		volumeMounts = append(volumeMounts, corev1.VolumeMount {Name: "certs-dir", MountPath: "/certs_dir"})

		// TODO: Shouldn't hard code /data here. It ties too closely with the
		// code in the backup container.
		env = append(env, corev1.EnvVar {Name: "CERTS_DEST_DIR", Value: "/data/certificates"})
		env = append(env, corev1.EnvVar {Name: "CERTS_SRC_DIR", Value: "/certs_dir"})
	}

	masterNodeLabelName := r.getMasterNodeLabelName(cr, log)

	return &batchv1beta1.CronJob {
		ObjectMeta: metav1.ObjectMeta {
			Name: cronJobName,
			Namespace: cr.Namespace,
		},

		Spec: batchv1beta1.CronJobSpec {
			ConcurrencyPolicy: "Forbid",
			Schedule: cr.Spec.Schedule,
			Suspend: cr.Spec.Suspend,

			JobTemplate: batchv1beta1.JobTemplateSpec {
				ObjectMeta: metav1.ObjectMeta {
					Name: cr.Name + "-backup-job",
					Namespace: cr.Namespace,
				},
				Spec: batchv1.JobSpec {
					// RD: Set backoffLimit to 2.
					Template: corev1.PodTemplateSpec {
						ObjectMeta: metav1.ObjectMeta {
							Name: cr.Name + "-backup-pod-template",
							Namespace: cr.Namespace,
						},
						Spec: corev1.PodSpec {
							RestartPolicy: "Never",
							HostNetwork: true,

							// To make sure that backup pod runs on the master.
							// TODO: We need to perhaps accept another label as configurable option.
							Affinity: &corev1.Affinity {
								NodeAffinity: &corev1.NodeAffinity {
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector {
										NodeSelectorTerms: [] corev1.NodeSelectorTerm {
											corev1.NodeSelectorTerm {
												MatchExpressions: [] corev1.NodeSelectorRequirement {
													corev1.NodeSelectorRequirement {
														Key: masterNodeLabelName,
														Operator: "Exists",
													},
												},
											},
										},
									},
								},
							},

							// Tolerate "NoSchedule" taint on master nodes.
							Tolerations: [] corev1.Toleration {
								corev1.Toleration {
									Operator: "Exists",
									Effect: "NoSchedule",
								},
							},
							
							Volumes: volumes,
							
							Containers: []corev1.Container {
								{
									Name: cr.Name + "-kcx-backup",
									Image:   "kubedrbackup:0.47",
									VolumeMounts: volumeMounts,
									Env: env,

									Args: []string {
										"sh", "/usr/local/bin/kubedrbackup.sh",
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
