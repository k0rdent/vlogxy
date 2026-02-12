package servergroup

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Group struct {
	// Target address:port list for promxy Prometheus server group static_configs
	Target string `yaml:"target"`
	// ClusterName is the promxyCluster label value
	ClusterName string `yaml:"cluster_name"`
	// PathPrefix defines path_prefix for all targets
	PathPrefix string `yaml:"path_prefix"`
	// Scheme for all targets (http or https)
	Scheme     string           `yaml:"scheme"`
	HttpClient HTTPClientConfig `yaml:"http_client"`
}

// HTTPClientConfig defines the http client TLS and BasicAuth config for Prometheus
type HTTPClientConfig struct {
	// DialTimeout in the string representation (e.g. 1s)
	DialTimeout metav1.Duration `yaml:"dial_timeout"`
	TLSConfig   TLSConfig       `yaml:"tls_config"`
	BasicAuth   BasicAuth       `yaml:"basic_auth"`
}

// BasicAuth part of prometheus HTTPClientConfig with yaml annotation
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// TLSConfig part of prometheus HTTPClientConfig with yaml annotation
type TLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}
