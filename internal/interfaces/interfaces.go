package interfaces

import (
	"context"
	"net/http"

	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
)

// ConfigProvider provides access to application configuration
type ConfigProvider interface {
	// GetServerGroups returns the list of server groups to query
	GetServerGroups() []*servergroup.Server
	// GetMaxLogsLimit returns the maximum number of logs to return in a query
	GetMaxLogsLimit() int
	// Reload reloads the configuration from source
	Reload() error
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
}

type StreamResponseAggregator[T any] interface {
	// StreamParseResponse processes a single backend response and sends data to the provided channel
	StreamParseResponse(context.Context, *http.Response, chan<- []byte)
	// GetMaxLogsLimit returns the maximum number of logs to return in a query
	GetMaxLogsLimit() int
	// GetBufferSize returns the buffer size for streaming responses
	GetBufferSize() int
}

type ProxyGroup[T any] interface {
	ProxyRequest()
}
