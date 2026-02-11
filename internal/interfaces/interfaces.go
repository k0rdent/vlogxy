package interfaces

import (
	"net/http"
)

// ConfigProvider provides access to application configuration
type ConfigProvider interface {
	// GetServerGroups returns the list of server groups to query
	GetServerGroups() []ServerGroup
	// Reload reloads the configuration from source
	Reload() error
}

// ServerGroup represents a VictoriaLogs backend server
type ServerGroup struct {
	Target string
	Scheme string
}

// HTTPClient defines interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResponseAggregator defines the interface for querying VictoriaLogs backends
type ResponseAggregator[T any] interface {
	// ParseResponse processes a single backend response
	ParseResponse(*http.Response) (T, error)
	// Merge combines responses from multiple backends
	Merge([]T) ([]byte, error)
	// GetURL returns the full URL for the query
	GetURL(scheme, target string) string
}
