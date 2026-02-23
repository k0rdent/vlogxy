package servergroup

import (
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/k0rdent/vlogxy/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Server struct {
	// Targets is a list of VictoriaLogs server endpoints (address:port) for high availability.
	// All endpoints are replicas with identical data. If a request fails on one,
	// the next target in the list will be tried.
	Targets []string `yaml:"targets"`
	// ClusterName is the cluster label value
	ClusterName string `yaml:"cluster_name"`
	// PathPrefix defines path_prefix for all targets
	PathPrefix string `yaml:"path_prefix"`
	// Scheme for all targets (http or https)
	Scheme string `yaml:"scheme"`
	// HttpClientConfig defines the HTTP client configuration for this server group
	HttpClientConfig HTTPClientConfig `yaml:"http_client"`

	httpClient *http.Client
	once       sync.Once
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

// URLs constructs the full URLs for all targets with a given path and query
func (s *Server) URLs(path, query string) []string {
	urls := make([]string, len(s.Targets))
	for i, target := range s.Targets {
		urls[i] = common.BuildURL(s.Scheme, target, s.PathPrefix+path, query)
	}
	return urls
}

// Username returns the basic auth username
func (s *Server) Username() string {
	return s.HttpClientConfig.BasicAuth.Username
}

// Password returns the basic auth password
func (s *Server) Password() string {
	return s.HttpClientConfig.BasicAuth.Password
}

// HTTPClient returns a configured HTTP client for this server group
func (s *Server) HTTPClient() *http.Client {
	s.once.Do(func() {
		s.httpClient = &http.Client{
			Timeout: s.HttpClientConfig.DialTimeout.Duration,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: s.HttpClientConfig.TLSConfig.InsecureSkipVerify,
				},
			},
		}
	})
	return s.httpClient
}
