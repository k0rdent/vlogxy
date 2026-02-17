package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

func newRequest(ctx context.Context, server *servergroup.Server, originalURL *url.URL) (*http.Response, error) {
	targetURL := server.URL(originalURL.Path, originalURL.RawQuery)
	httpClient := server.HTTPClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", targetURL, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(server.Username(), server.Password())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %w", server.Target, err)
	}

	return resp, nil
}

func collectResponses(ctx context.Context, serverGroup []*servergroup.Server, originalURL *url.URL) <-chan *http.Response {
	wg := &sync.WaitGroup{}
	ch := make(chan *http.Response, len(serverGroup))

	for i := range serverGroup {
		wg.Go(func() {
			server := serverGroup[i]
			resp, err := newRequest(ctx, server, originalURL)
			if err != nil {
				log.Errorf("failed to make request to %s: %v", server.Target, err)
				return
			}

			select {
			case <-ctx.Done():
				return
			case ch <- resp:
			}
		})
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
