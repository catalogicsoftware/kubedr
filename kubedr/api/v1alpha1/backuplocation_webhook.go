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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var backuplocationlog = logf.Log.WithName("backuplocation-resource")

// SetupWebhookWithManager configures the web hook with the manager.
func (r *BackupLocation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-kubedr-catalogicsoftware-com-v1alpha1-backuplocation,mutating=true,failurePolicy=fail,groups=kubedr.catalogicsoftware.com,resources=backuplocations,verbs=create;update,versions=v1alpha1,name=mbackuplocation.kb.io

var _ webhook.Defaulter = &BackupLocation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *BackupLocation) Default() {
	backuplocationlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-kubedr-catalogicsoftware-com-v1alpha1-backuplocation,mutating=false,failurePolicy=fail,groups=kubedr.catalogicsoftware.com,resources=backuplocations,versions=v1alpha1,name=vbackuplocation.kb.io

var _ webhook.Validator = &BackupLocation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *BackupLocation) ValidateCreate() error {
	backuplocationlog.Info("validate create", "name", r.Name)

	// We need to validate that Credentials are correct.
	//
	// The quickest way for now is to actually try and initialize the repo.
	// The command will fail if credentials are wrong.

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *BackupLocation) ValidateUpdate(old runtime.Object) error {
	backuplocationlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *BackupLocation) ValidateDelete() error {
	backuplocationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
