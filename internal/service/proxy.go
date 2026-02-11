package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/k0rdent/victorialogs-aggregator/internal/interfaces"
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

			select {
			case <-ctx.Done():
				log.Warnf("context cancelled, skipping request to %s", group.Target)
				return
			default:
			}

			endpoint := querier.GetURL(group.Scheme, group.Target)
			response, err := s.forwardRequest(ctx, endpoint)
			if err != nil {
				log.Errorf("failed to forward request to %s: %v", endpoint, err)
				return
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from %s", response.StatusCode, endpoint)
				return
			}

			resp, err := querier.ParseResponse(response)
			if err != nil {
				log.Errorf("failed to handle response from %s: %v", endpoint, err)
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

func (s *ProxyService) forwardRequest(ctx context.Context, endpoint string) (*http.Response, error) {
	proxyReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Errorf("failed to create request for %s: %v", endpoint, err)
		return nil, err
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		log.Errorf("failed to forward request to %s: %v", endpoint, err)
		return nil, err
	}

	return proxyResp, nil
}
