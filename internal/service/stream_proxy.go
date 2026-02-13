package service

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

const bufferSize = 5

func Stream[T any](ctx context.Context, s *ProxyService, querier interfaces.StreamResponseAggregator[T]) <-chan []byte {
	serverGroups := s.config.GetServerGroups()
	resultChan := make(chan []byte)

	go func() {
		defer close(resultChan)

		respChans := s.collectStreamingResponses(ctx, serverGroups, querier)
		if len(respChans) == 0 {
			log.Warn("no successful streaming responses from backends")
			return
		}

		s.mergeAndStreamSortedData(ctx, respChans, resultChan)
	}()

	return resultChan
}

// collectStreamingResponses fetches streaming responses from all backends concurrently
func (s *ProxyService) collectStreamingResponses(
	ctx context.Context,
	serverGroups []servergroup.Group,
	querier interfaces.StreamResponseAggregator[any],
) []<-chan []byte {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	respChans := make([]<-chan []byte, 0, len(serverGroups))

	for i := range serverGroups {
		wg.Go(func() {
			group := serverGroups[i]
			url := querier.GetURL(group.Scheme, group.Target, group.PathPrefix)

			response, err := s.forwardRequest(ctx, group, url)
			if err != nil {
				log.Errorf("failed to forward request to %s: %v", group.Target, err)
				return
			}

			if response.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from %s", response.StatusCode, group.Target)
				if err := response.Body.Close(); err != nil {
					log.Errorf("failed to close response body: %v", err)
				}
				return
			}

			respChan, err := querier.StreamParseResponse(ctx, response)
			if err != nil {
				log.Errorf("failed to parse streaming response from %s: %v", group.Target, err)
				return
			}

			mu.Lock()
			respChans = append(respChans, respChan)
			mu.Unlock()
		})
	}

	wg.Wait()
	return respChans
}

// mergeAndStreamSortedData merges data from multiple channels and streams sorted results
func (s *ProxyService) mergeAndStreamSortedData(
	ctx context.Context,
	respChans []<-chan []byte,
	resultChan chan<- []byte,
) {
	dataChan := make(chan []byte)
	var wg sync.WaitGroup

	for i := range respChans {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case data, ok := <-respChans[i]:
					if !ok {
						return
					}
					dataChan <- data
				}
			}
		})
	}

	// Close dataChan when all sources are done
	go func() {
		wg.Wait()
		close(dataChan)
	}()

	buffer := make([]map[string]any, 0, bufferSize)

	for data := range dataChan {
		var item map[string]any
		if err := json.Unmarshal(data, &item); err != nil {
			log.Errorf("failed to unmarshal streaming data: %v", err)
			continue
		}

		buffer = append(buffer, item)

		if len(buffer) == bufferSize {
			if err := s.flushSortedBuffer(buffer, resultChan); err != nil {
				log.Errorf("failed to flush buffer: %v", err)
			}
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		if err := s.flushSortedBuffer(buffer, resultChan); err != nil {
			log.Errorf("failed to flush remaining buffer: %v", err)
		}
	}
}

// flushSortedBuffer sorts buffer by timestamp and sends to result channel
func (s *ProxyService) flushSortedBuffer(
	buffer []map[string]any,
	resultChan chan<- []byte,
) error {
	sort.Slice(buffer, func(i, j int) bool {
		tsI, okI := buffer[i]["_time"].(float64)
		tsJ, okJ := buffer[j]["_time"].(float64)
		if !okI || !okJ {
			return false
		}
		return tsI < tsJ
	})

	for _, item := range buffer {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		resultChan <- data
	}

	return nil
}
