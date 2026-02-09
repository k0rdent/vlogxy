package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VLServerGroupSpec defines the desired state of PromxyServerGroup
type VLServerGroupSpec struct {
	// ClusterName is the promxyCluster label value
	ClusterName string `json:"cluster_name,omitempty"`
	// Targets address:port list for promxy Prometheus server group static_configs
	Targets []string `json:"targets,omitempty"`
	// PathPrefix defines path_prefix for all targets
	PathPrefix string `json:"path_prefix,omitempty"`
	// Scheme for all targets (http or https)
	Scheme     string           `json:"scheme,omitempty"`
	HttpClient HTTPClientConfig `json:"http_client,omitempty"`
}

// HTTPClientConfig defines the http client TLS and BasicAuth config for Prometheus
type HTTPClientConfig struct {
	// DialTimeout in the string representation (e.g. 1s)
	DialTimeout metav1.Duration `json:"dial_timeout,omitempty"`
	TLSConfig   TLSConfig       `json:"tls_config,omitempty"`
	BasicAuth   BasicAuth       `json:"basic_auth,omitempty"`
}

// BasicAuth part of prometheus HTTPClientConfig with json annotation
type BasicAuth struct {
	CredentialsSecretName string `json:"credentials_secret_name,omitempty"`
	UsernameKey           string `json:"username_key,omitempty"`
	PasswordKey           string `json:"password_key,omitempty"`
}

// TLSConfig part of prometheus HTTPClientConfig with json annotation
type TLSConfig struct {
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`
}

// VLServerGroupStatus defines the observed state of VLServerGroup
type VLServerGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VLServerGroup is the Schema for the vlservergroups API
type VLServerGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VLServerGroupSpec   `json:"spec,omitempty"`
	Status VLServerGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VLServerGroupList contains a list of VLServerGroup
type VLServerGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VLServerGroup `json:"items"`
}

// func init() {
// 	SchemeBuilder.Register(&VLServerGroup{}, &VLServerGroupList{})
// }
