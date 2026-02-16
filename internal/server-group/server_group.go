package servergroup

import (
	"crypto/tls"
	"net/http"

	"github.com/k0rdent/vlogxy/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Server struct {
	// Target address:port for the VictoriaLogs server
	Target string `yaml:"target"`
	// ClusterName is the cluster label value
	ClusterName string `yaml:"cluster_name"`
	// PathPrefix defines path_prefix for all targets
	PathPrefix string `yaml:"path_prefix"`
	// Scheme for all targets (http or https)
	Scheme string `yaml:"scheme"`
	// HttpClientConfig defines the HTTP client configuration for this server group
	HttpClientConfig HTTPClientConfig `yaml:"http_client"`
}

// HTTPClientConfig defines the http client TLS and BasicAuth config
type HTTPClientConfig struct {
	// DialTimeout in the string representation (e.g. 1s)
	DialTimeout metav1.Duration `yaml:"dial_timeout"`
	TLSConfig   TLSConfig       `yaml:"tls_config"`
	BasicAuth   BasicAuth       `yaml:"basic_auth"`
}

// BasicAuth credentials for HTTP authentication
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// TLSConfig for HTTPS connections
type TLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// URL constructs the full URL for a given path and query
func (s Server) URL(path, query string) string {
	return common.BuildURL(s.Scheme, s.Target, s.PathPrefix+path, query)
}

// Username returns the basic auth username
func (s Server) Username() string {
	return s.HttpClientConfig.BasicAuth.Username
}

// Password returns the basic auth password
func (s Server) Password() string {
	return s.HttpClientConfig.BasicAuth.Password
}

func (s Server) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: s.HttpClientConfig.DialTimeout.Duration,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: s.HttpClientConfig.TLSConfig.InsecureSkipVerify,
			},
		},
	}
}
