package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"

	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

func newRequest(ctx context.Context, server *servergroup.Server, method, path, rawQuery string, bodyBytes []byte) (*http.Response, error) {
	var lastErr error
	targetURLs := server.URLs(path, rawQuery)

	rand.Shuffle(len(targetURLs), func(i, j int) {
		targetURLs[i], targetURLs[j] = targetURLs[j], targetURLs[i]
	})

	for _, targetURL := range targetURLs {
		log.Debugf("Requesting data from target: %s", targetURL)

		httpClient := server.HTTPClient()
		var bodyReader io.Reader
		if method == http.MethodPost && len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request for %s: %w", targetURL, err)
			log.Warnf("could not create request: %v, trying next target", lastErr)
			continue
		}

		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
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

func collectResponses(ctx context.Context, serverGroup []*servergroup.Server, originalReq *http.Request) <-chan *http.Response {
	var bodyBytes []byte
	if originalReq.Method == http.MethodPost && originalReq.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(originalReq.Body)
		if err != nil {
			log.Errorf("failed to read request body: %v", err)
		}
	}

	wg := &sync.WaitGroup{}
	ch := make(chan *http.Response, len(serverGroup))

	for i := range serverGroup {
		wg.Go(func() {
			server := serverGroup[i]
			resp, err := newRequest(ctx, server, originalReq.Method, originalReq.URL.Path, originalReq.URL.RawQuery, bodyBytes)
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
