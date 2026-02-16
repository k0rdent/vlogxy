package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

const bufferSize = 5

type StreamProxyGroup[T any] interface {
	ProxyRequest(interfaces.StreamResponseAggregator[T])
}

type StreamProxy[T any] struct {
	serverGroup []servergroup.Server
	httpClient  interfaces.HTTPClient
	ginContext  *gin.Context
	limit       int
}

func NewStreamProxy[T any](serverGroup []servergroup.Server, httpClient interfaces.HTTPClient, c *gin.Context) StreamProxyGroup[T] {
	return &StreamProxy[T]{
		serverGroup: serverGroup,
		httpClient:  httpClient,
		limit:       getLimit(c),
		ginContext:  c,
	}
}

func (s *StreamProxy[T]) ProxyRequest(aggregator interfaces.StreamResponseAggregator[T]) {
	ctx := s.ginContext.Request.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	respChan := collectResponses(ctx, s.serverGroup, s.ginContext.Request.URL)
	dataChan := make(chan []byte)
	wg := sync.WaitGroup{}

	for resp := range respChan {
		wg.Go(func() {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Errorf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from backend", resp.StatusCode)
				return
			}

			aggregator.StreamParseResponse(ctx, resp, dataChan)
		})
	}

	go func() {
		wg.Wait()
		close(dataChan)
	}()

	buffer := make([]map[string]any, 0, bufferSize)
	remainingLimit := s.limit

	s.ginContext.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			log.Warn("client disconnected, stopping stream")
			return false
		case output, ok := <-dataChan:
			if !ok {
				return false
			}

			var item map[string]any
			if err := json.Unmarshal(output, &item); err != nil {
				log.Errorf("failed to unmarshal streaming data: %v", err)
				return false
			}

			buffer = append(buffer, item)

			if len(buffer) == bufferSize || (remainingLimit > 0 && len(buffer) >= remainingLimit) {
				s.sortBuffer(buffer)

				for _, sortedItem := range buffer {
					sortedData, err := json.Marshal(sortedItem)
					if err != nil {
						log.Errorf("failed to marshal sorted item: %v", err)
						continue
					}
					if _, err := w.Write(sortedData); err != nil {
						log.Errorf("failed to write sorted data: %v", err)
						return false
					}
					if _, err := w.Write([]byte("\n")); err != nil {
						log.Errorf("failed to write newline: %v", err)
						return false
					}
				}

				remainingLimit -= len(buffer)
				buffer = buffer[:0]

				if remainingLimit <= 0 && s.limit > 0 {
					return false
				}
			}

			return true
		}
	})

	if len(buffer) > 0 {
		s.sortBuffer(buffer)

		w := s.ginContext.Writer
		for _, sortedItem := range buffer {
			sortedData, err := json.Marshal(sortedItem)
			if err != nil {
				log.Errorf("failed to marshal sorted item: %v", err)
				continue
			}
			if _, err := w.Write(sortedData); err != nil {
				log.Errorf("failed to write sorted data: %v", err)
				break
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				log.Errorf("failed to write newline: %v", err)
				break
			}
		}
	}
}

func (s *StreamProxy[T]) sortBuffer(buffer []map[string]any) {
	sort.Slice(buffer, func(i, j int) bool {
		tsI, okI := buffer[i]["_time"].(float64)
		tsJ, okJ := buffer[j]["_time"].(float64)
		if !okI || !okJ {
			return false
		}
		return tsI < tsJ
	})
}

func getLimit(c *gin.Context) int {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 0
	}
	return limit
}
