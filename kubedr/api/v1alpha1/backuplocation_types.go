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

// BackupLocationSpec defines the desired state of BackupLocation
type BackupLocationSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// kubebuilder:validation:MinLength:=1
	Url string `json:"url"`
	// kubebuilder:validation:MinLength:=1
	BucketName string `json:"bucketName"`

	// name of the secret
	// kubebuilder:validation:MinLength:=1
	Credentials string `json:"credentials"`
}

// BackupLocationStatus defines the observed state of BackupLocation
type BackupLocationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	InitStatus string `json:"initStatus,omitempty"`

	// +kubebuilder:validation:Optional
	InitErrorMessage string `json:"initErrorMessage,omitempty"`

	InitTime string `json:"initTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BackupLocation is the Schema for the backuplocations API
type BackupLocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupLocationSpec   `json:"spec,omitempty"`
	Status BackupLocationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupLocationList contains a list of BackupLocation
type BackupLocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupLocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupLocation{}, &BackupLocationList{})
}
