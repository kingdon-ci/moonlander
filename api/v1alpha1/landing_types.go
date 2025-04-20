package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LandingSpec defines the desired state of Landing
type LandingSpec struct {
	KubeconfigSecretName         string `json:"kubeconfigSecretName,omitempty"`
	KubeconfigSecretNamespace    string `json:"kubeconfigSecretNamespace,omitempty"`
	WriteKubeconfigSecretName    string `json:"writeKubeconfigSecretName,omitempty"`
	TargetNamespace              string `json:"targetNamespace,omitempty"`
}

// LandingStatus defines the observed state of Landing
type LandingStatus struct {
	// Add status fields if needed
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Landing is the Schema for the landings API
type Landing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LandingSpec   `json:"spec,omitempty"`
	Status LandingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LandingList contains a list of Landing
type LandingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Landing `json:"items"`
}
