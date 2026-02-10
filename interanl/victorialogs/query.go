package victorialogs

import (
	"net/http"
	"sync"

	"github.com/k0rdent/victorialogs-aggregator/interanl/config"
	log "github.com/sirupsen/logrus"
)

type Querier[T any] interface {
	Merge([]T) ([]byte, error)
	ResponseHandler(*http.Response) (T, error)
	GetEndpoint(scheme, target string) string
}

func MultiClusterProxy[T any](req Querier[T]) ([]byte, error) {
	var wg sync.WaitGroup
	var mergedLogs []T
	configData := config.GlobalConfig.GetData()
	reader := make(chan T, len(configData.ServerGroups))

	var readerWg sync.WaitGroup
	readerWg.Go(func() {
		for resp := range reader {
			mergedLogs = append(mergedLogs, resp)
		}
	})

	for i := range configData.ServerGroups {
		sg := configData.ServerGroups[i]
		wg.Go(func() {
			endpoint := req.GetEndpoint(sg.Scheme, sg.Target)
			response, err := forwardRequest(endpoint)
			if err != nil {
				log.Errorf("failed to forward request to %s: %v", endpoint, err)
				return
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from %s", response.StatusCode, endpoint)
				return
			}

			resp, err := req.ResponseHandler(response)
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

	return req.Merge(mergedLogs)
}

func forwardRequest(endpoint string) (*http.Response, error) {
	proxyReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Errorf("failed to create request for %s: %v", endpoint, err)
		return nil, err
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		log.Errorf("failed to forward request to %s: %v", endpoint, err)
		return nil, err
	}

	return proxyResp, nil
}
