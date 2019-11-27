package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetadataBackupPolicySpec defines the desired state of MetadataBackupPolicy
// +k8s:openapi-gen=true
type MetadataBackupPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Name of the S3 BackupLocation resource
	Destination string `json:"destination"`

	// Optional. If not provided, certificates will not be backed up.

	// +optional
    CertsDir string `json:"certsDir"`

    EtcdEndpoint string `json:"etcdEndpoint"`

	// Name of the "secret" containing etcd certificates.
	EtcdCreds string `json:"etcdCreds"`

	// The value of this field should be either "now" or
	// same as "schedule" in "cronjob".
	Schedule string `json:"schedule"`
}

// MetadataBackupPolicyStatus defines the observed state of MetadataBackupPolicy
// +k8s:openapi-gen=true
type MetadataBackupPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetadataBackupPolicy is the Schema for the metadatabackuppolicies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MetadataBackupPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetadataBackupPolicySpec   `json:"spec,omitempty"`
	Status MetadataBackupPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetadataBackupPolicyList contains a list of MetadataBackupPolicy
type MetadataBackupPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetadataBackupPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetadataBackupPolicy{}, &MetadataBackupPolicyList{})
}
