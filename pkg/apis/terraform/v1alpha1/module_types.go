package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:subresource:status
// ModuleSpec defines the desired state of Module
type ModuleSpec struct {
	// Kubernetes namespace module source
	// +kubebuilder:validation:Enum=/var/lib/modules/kubernetes/namespace/
	Source string `json:"source,omitempty"`
	// Kubernetes namespace name
	NamespaceName string `json:"namespace_name,omitempty"`
}

// +genclient
// +genclient:Namespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=modules,singular=module,scope=Namespaced,shortName=module
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status",description="Description of the current status"
// Module is the Schema for the modules API
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleSpec `json:"spec,omitempty"`
	Status string     `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// ModuleList contains a list of Module
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Module `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Module{}, &ModuleList{})
}
