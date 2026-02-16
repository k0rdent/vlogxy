package servergroup

import (
	"crypto/tls"
	"net/http"

	"github.com/k0rdent/vlogxy/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Server struct {
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

func (s Server) URL(path, query string) string {
	return common.BuildURL(s.Scheme, s.Target, s.PathPrefix+path, query)
}

func (s Server) Username() string {
	return s.HttpClient.BasicAuth.Username
}

func (s Server) Password() string {
	return s.HttpClient.BasicAuth.Password
}

func (s Server) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: s.HttpClient.DialTimeout.Duration,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: s.HttpClient.TLSConfig.InsecureSkipVerify,
			},
		},
	}
}
