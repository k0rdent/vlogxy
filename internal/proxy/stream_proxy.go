package proxy

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

const bufferSize = 50

type StreamProxyGroup[T any] interface {
	ProxyRequest(interfaces.StreamResponseAggregator[T])
}

type StreamProxy[T any] struct {
	serverGroup []*servergroup.Server
	httpClient  interfaces.HTTPClient
	ginContext  *gin.Context
	limit       int
}

func NewStreamProxy[T any](config interfaces.ConfigProvider, httpClient interfaces.HTTPClient, c *gin.Context) StreamProxyGroup[T] {
	return &StreamProxy[T]{
		serverGroup: config.GetServerGroups(),
		httpClient:  httpClient,
		limit:       getLimit(c, config.GetMaxLogsLimit()),
		ginContext:  c,
	}
}

func (s *StreamProxy[T]) ProxyRequest(aggregator interfaces.StreamResponseAggregator[T]) {
	ctx := s.ginContext.Request.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dataChan := s.collectStreamData(ctx, aggregator)
	s.streamToClient(ctx, dataChan)
}

// collectStreamData spawns goroutines to collect data from all backends
func (s *StreamProxy[T]) collectStreamData(ctx context.Context, aggregator interfaces.StreamResponseAggregator[T]) <-chan []byte {
	respChan := collectResponses(ctx, s.serverGroup, s.ginContext.Request.URL)
	dataChan := make(chan []byte, 100)
	wg := &sync.WaitGroup{}

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

	return dataChan
}

// streamToClient reads from dataChan and streams sorted batches to the client
func (s *StreamProxy[T]) streamToClient(ctx context.Context, dataChan <-chan []byte) {
	buffer := make([]map[string]any, 0, bufferSize)
	remainingLimit := s.limit

	s.ginContext.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			log.Warn("client disconnected, stopping stream")
			return false
		case data, ok := <-dataChan:
			if !ok {
				s.flushBuffer(w, buffer)
				return false
			}

			item, err := s.unmarshalItem(data)
			if err != nil {
				log.Errorf("failed to unmarshal streaming data: %v", err)
				return false
			}

			buffer = append(buffer, item)

			if s.shouldFlushBuffer(buffer, remainingLimit) {
				if !s.writeBuffer(w, buffer) {
					return false
				}
				remainingLimit -= len(buffer)
				buffer = buffer[:0]

				if s.limitReached(remainingLimit) {
					return false
				}
			}

			return true
		}
	})
}

// unmarshalItem unmarshals JSON data into a map
func (s *StreamProxy[T]) unmarshalItem(data []byte) (map[string]any, error) {
	var item map[string]any
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return item, nil
}

// shouldFlushBuffer determines if the buffer should be flushed
func (s *StreamProxy[T]) shouldFlushBuffer(buffer []map[string]any, remainingLimit int) bool {
	return len(buffer) >= bufferSize || (remainingLimit > 0 && len(buffer) >= remainingLimit)
}

// limitReached checks if the output limit has been reached
func (s *StreamProxy[T]) limitReached(remainingLimit int) bool {
	return remainingLimit <= 0 && s.limit > 0
}

// writeBuffer sorts and writes the buffer to the writer
func (s *StreamProxy[T]) writeBuffer(w io.Writer, buffer []map[string]any) bool {
	s.sortBuffer(buffer)

	for _, item := range buffer {
		if !s.writeItem(w, item) {
			return false
		}
	}
	return true
}

// flushBuffer writes any remaining items in the buffer
func (s *StreamProxy[T]) flushBuffer(w io.Writer, buffer []map[string]any) {
	if len(buffer) > 0 {
		s.writeBuffer(w, buffer)
	}
}

// writeItem marshals and writes a single item to the writer
func (s *StreamProxy[T]) writeItem(w io.Writer, item map[string]any) bool {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(item); err != nil {
		log.Errorf("failed to encode item: %v", err)
		return false
	}
	return true
}

func (s *StreamProxy[T]) sortBuffer(buffer []map[string]any) {
	slices.SortStableFunc(buffer, func(a, b map[string]any) int {
		tsA := a["_time"].(string)
		tsB := b["_time"].(string)
		timeA, err := time.Parse(time.RFC3339Nano, tsA)
		if err != nil {
			log.Errorf("failed to parse timestamp: %v", err)
			return 0
		}
		timeB, err := time.Parse(time.RFC3339Nano, tsB)
		if err != nil {
			log.Errorf("failed to parse timestamp: %v", err)
			return 0
		}
		return cmp.Compare(timeB.Unix(), timeA.Unix())
	})
}

func getLimit(c *gin.Context, maxLimit int) int {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return maxLimit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return maxLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
