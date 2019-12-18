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

package v1alpha1

import (
	"github.com/robfig/cron"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// log is for logging in this package.
var log = logf.Log.WithName("metadatabackuppolicy-resource")

func (r *MetadataBackupPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:verbs=create;update,path=/mutate-kubedr-catalogicsoftware-com-v1alpha1-metadatabackuppolicy,mutating=true,failurePolicy=fail,groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies,versions=v1alpha1,name=mutatemetadatabackuppolicy.kb.io

var _ webhook.Defaulter = &MetadataBackupPolicy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MetadataBackupPolicy) Default() {
	log.Info("default", "name", r.Name)

	if r.Spec.EtcdEndpoint == "" {
		log.Info("Initializing EtcdEndpoint")
		r.Spec.EtcdEndpoint = "https://127.0.0.1:2379"
	}

	if r.Spec.EtcdCreds == "" {
		log.Info("Initializing EtcdCreds")
		r.Spec.EtcdCreds = "etcd-creds"
	}

	if r.Spec.RetainNumBackups == nil || *r.Spec.RetainNumBackups == 0 {
		log.Info("Initializing RetainNumBackups")
		r.Spec.RetainNumBackups = new(int64)
		*r.Spec.RetainNumBackups = 120
	}

	if r.Spec.Suspend == nil {
		log.Info("Initializing 'Suspend'")
		// Initialized to false.
		r.Spec.Suspend = new(bool)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-kubedr-catalogicsoftware-com-v1alpha1-metadatabackuppolicy,mutating=false,failurePolicy=fail,groups=kubedr.catalogicsoftware.com,resources=metadatabackuppolicies,versions=v1alpha1,name=vmetadatabackuppolicy.kb.io

var _ webhook.Validator = &MetadataBackupPolicy{}

func validateScheduleFormat(schedule string, fldPath *field.Path) *field.Error {
	if _, err := cron.ParseStandard(schedule); err != nil {
		return field.Invalid(fldPath, schedule, err.Error())
	}
	return nil
}

func (r *MetadataBackupPolicy) validateCronJobSpec() *field.Error {
	return validateScheduleFormat(
		r.Spec.Schedule,
		field.NewPath("spec").Child("schedule"))
}

func (r *MetadataBackupPolicy) validatePolicy() error {
	var allErrs field.ErrorList

	if err := r.validateCronJobSpec(); err != nil {
		allErrs = append(allErrs, err)
	}

	// Validate destination
	// Verify that the resource exists.

	// Validate etcd endpoint and creds
	// Connect and issue a dummy command

	// TODO: How to validate certs dir?

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "kubedr.catalogicsoftware.com/v1alpha1", Kind: "MetadataBackupPolicy"},
		r.Name, allErrs)
}


// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MetadataBackupPolicy) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)

	return r.validatePolicy()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MetadataBackupPolicy) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)
	return r.validatePolicy()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MetadataBackupPolicy) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
