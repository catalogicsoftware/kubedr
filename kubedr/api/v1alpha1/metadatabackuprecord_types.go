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

// MetadataBackupRecordSpec defines the desired state of MetadataBackupRecord
type MetadataBackupRecordSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// kubebuilder:validation:MinLength:=1
    SnapshotId string `json:"snapshotId"`

	// kubebuilder:validation:MinLength:=1
    Policy string `json:"policy"`
}

// MetadataBackupRecordStatus defines the observed state of MetadataBackupRecord
type MetadataBackupRecordStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// MetadataBackupRecord is the Schema for the metadatabackuprecords API
type MetadataBackupRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetadataBackupRecordSpec   `json:"spec,omitempty"`
	Status MetadataBackupRecordStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MetadataBackupRecordList contains a list of MetadataBackupRecord
type MetadataBackupRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetadataBackupRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetadataBackupRecord{}, &MetadataBackupRecordList{})
}
