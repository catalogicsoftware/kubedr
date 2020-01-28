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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetadataBackupPolicySpec defines the desired state of MetadataBackupPolicy
type MetadataBackupPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the S3 BackupLocation resource
	// kubebuilder:validation:MinLength:=1
	Destination string `json:"destination"`

	// Optional. If not provided, certificates will not be backed up.
	// +kubebuilder:validation:Optional
	CertsDir string `json:"certsDir,omitempty"`

	// +kubebuilder:validation:Optional
	// If not provided, "https://127.0.0.1:2379" will be used.
	EtcdEndpoint string `json:"etcdEndpoint,omitempty"`

	// Name of the "secret" containing etcd certificates.
	// +kubebuilder:validation:Optional
	// If not provided, "etcd-creds" is used as the name of the secret comprising of
	// credentials.
	EtcdCreds string `json:"etcdCreds,omitempty"`

	// The value of this field should be same as "schedule" in "cronjob".
	Schedule string `json:"schedule"`

	// Refers to name of a configmap containing list of key=value pairs.
	// Options string `json:"options"`
	// +kubebuilder:validation:Optional
	Options map[string]string `json:"options,omitempty"`

	// Props map[string]string `json:"props"`

	// Should we even have default?
	// +kubebuilder:validation:Optional
	RetainNumBackups *int64 `json:"retainNumBackups,omitempty"`

	// +kubebuilder:validation:Optional
	Suspend *bool `json:"suspend,omitempty"`
}

// MetadataBackupPolicyStatus defines the observed state of MetadataBackupPolicy
type MetadataBackupPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// MetadataBackupPolicy is the Schema for the metadatabackuppolicies API
type MetadataBackupPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetadataBackupPolicySpec   `json:"spec,omitempty"`
	Status MetadataBackupPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MetadataBackupPolicyList contains a list of MetadataBackupPolicy
type MetadataBackupPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetadataBackupPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetadataBackupPolicy{}, &MetadataBackupPolicyList{})
}
