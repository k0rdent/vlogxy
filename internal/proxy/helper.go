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
	var lastErr error
	targetURLs := server.URLs(originalURL.Path, originalURL.RawQuery)

	for _, targetURL := range targetURLs {
		log.Debugf("Requesting data from target: %s", targetURL)

		httpClient := server.HTTPClient()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request for %s: %w", targetURL, err)
			log.Warnf("could not create request: %v, trying next target", lastErr)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(server.Username(), server.Password())

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", targetURL, err)
			log.Warnf("request failed: %v, trying next target", lastErr)
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("no targets available for server")
}

func collectResponses(ctx context.Context, serverGroup []*servergroup.Server, originalURL *url.URL) <-chan *http.Response {
	wg := &sync.WaitGroup{}
	ch := make(chan *http.Response, len(serverGroup))

	for i := range serverGroup {
		wg.Go(func() {
			server := serverGroup[i]
			resp, err := newRequest(ctx, server, originalURL)
			if err != nil {
				log.Errorf("failed to make request to %s: %v", server.ClusterName, err)
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
