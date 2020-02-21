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

// MetadataRestoreSpec defines the desired state of MetadataRestore
type MetadataRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// kubebuilder:validation:MinLength:=1
	MBRName string `json:"mbrName"`

	// kubebuilder:validation:MinLength:=1
	PVCName string `json:"pvcName"`
}

// MetadataRestoreStatus defines the observed state of MetadataRestore
type MetadataRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration"`

	RestoreStatus string `json:"restoreStatus"`

	// +kubebuilder:validation:Optional
	RestoreErrorMessage string `json:"restoreErrorMessage"`

	RestoreTime string `json:"restoreTime"`
}

// The creation of this resource triggers full restore of the data
// (etcd snapshot and certificates (if they were part of the backup).
// It would have been ideal to use a custom subresource (such as
// "/restore" but custom subresources are not yet supported for
// custom resources.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MetadataRestore is the Schema for the metadatarestores API
type MetadataRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetadataRestoreSpec   `json:"spec,omitempty"`
	Status MetadataRestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MetadataRestoreList contains a list of MetadataRestore
type MetadataRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetadataRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetadataRestore{}, &MetadataRestoreList{})
}
