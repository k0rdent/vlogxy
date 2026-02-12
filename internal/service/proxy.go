package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

// ProxyService implements multi-cluster proxy operations
type ProxyService struct {
	config     interfaces.ConfigProvider
	httpClient interfaces.HTTPClient
}

// NewProxyService creates a new proxy service instance
func NewProxyService(config interfaces.ConfigProvider, httpClient interfaces.HTTPClient) *ProxyService {
	return &ProxyService{
		config:     config,
		httpClient: httpClient,
	}
}

// Execute performs a query across all configured backends
func Execute[T any](ctx context.Context, s *ProxyService, querier interfaces.ResponseAggregator[T]) ([]byte, error) {
	serverGroups := s.config.GetServerGroups()

	var wg sync.WaitGroup
	var mergedLogs []T
	reader := make(chan T, len(serverGroups))

	var readerWg sync.WaitGroup
	readerWg.Go(func() {
		for resp := range reader {
			mergedLogs = append(mergedLogs, resp)
		}
	})

	for i := range serverGroups {
		wg.Go(func() {
			group := serverGroups[i]

			url := querier.GetURL(group.Scheme, group.Target, group.PathPrefix)
			response, err := s.forwardRequest(ctx, group, url)
			if err != nil {
				log.Errorf("failed to forward request to %s: %v", group.Target, err)
				return
			}
			defer func() {
				if err := response.Body.Close(); err != nil {
					log.Errorf("failed to close response body from %s: %v", group.Target, err)
				}
			}()

			if response.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from %s", response.StatusCode, group.Target)
				return
			}

			resp, err := querier.ParseResponse(response)
			if err != nil {
				log.Errorf("failed to handle response from %s: %v", group.Target, err)
				return
			}

			reader <- resp
		})
	}

	wg.Wait()
	close(reader)
	readerWg.Wait()

	return querier.Merge(mergedLogs)
}

func (s *ProxyService) forwardRequest(ctx context.Context, group servergroup.Group, url string) (*http.Response, error) {
	username := group.HttpClient.BasicAuth.Username
	password := group.HttpClient.BasicAuth.Password

	proxyReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("failed to create request for %s: %v", url, err)
		return nil, err
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.SetBasicAuth(username, password)

	proxyResp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		log.Errorf("failed to forward request to %s: %v", url, err)
		return nil, err
	}

	return proxyResp, nil
}
