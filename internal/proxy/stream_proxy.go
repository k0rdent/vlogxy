package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

const flushInterval = 1 * time.Second

type StreamProxyGroup[T any] interface {
	ProxyRequest()
}

type StreamProxy[T any] struct {
	serverGroup []*servergroup.Server
	httpClient  interfaces.HTTPClient
	aggregator  interfaces.StreamResponseAggregator[T]
	ginContext  *gin.Context
	limit       int
}

func NewStreamProxy[T any](serverGroup []*servergroup.Server, httpClient interfaces.HTTPClient, c *gin.Context, aggregator interfaces.StreamResponseAggregator[T]) StreamProxyGroup[T] {
	return &StreamProxy[T]{
		ginContext:  c,
		aggregator:  aggregator,
		httpClient:  httpClient,
		serverGroup: serverGroup,
		limit:       getLimit(c, aggregator.GetMaxLogsLimit()),
	}
}

func (s *StreamProxy[T]) ProxyRequest() {
	ctx := s.ginContext.Request.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dataChan := s.collectStreamData(ctx)
	s.streamToClient(ctx, dataChan)
}

// collectStreamData spawns goroutines to collect data from all backends
func (s *StreamProxy[T]) collectStreamData(ctx context.Context) <-chan []byte {
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

			s.aggregator.StreamParseResponse(ctx, resp, dataChan)
		})
	}

	go func() {
		wg.Wait()
		close(dataChan)
	}()

	return dataChan
}

// streamToClient reads from dataChan and streams sorted batches to the client.
func (s *StreamProxy[T]) streamToClient(ctx context.Context, dataChan <-chan []byte) {
	buffer := NewLogsBuffer(s.aggregator.GetBufferSize())
	remainingLimit := s.limit

	flushTimer := time.NewTimer(flushInterval)
	defer flushTimer.Stop()

	s.ginContext.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			log.Info("client disconnected, stopping stream")
			return false

		case data, ok := <-dataChan:
			if !ok {
				if err := buffer.Write(w); err != nil {
					log.Errorf("failed to write remaining buffer: %v", err)
				}
				return false
			}

			item, err := s.unmarshalItem(data)
			if err != nil {
				log.Errorf("failed to unmarshal streaming data: %v", err)
				return false
			}

			buffer.AddLog(item)

			if s.shouldFlushBuffer(buffer, remainingLimit) {
				return s.flushBuffer(w, buffer, &remainingLimit)
			}

			return true

		case <-flushTimer.C:
			ok := s.flushBuffer(w, buffer, &remainingLimit)
			flushTimer.Reset(flushInterval)
			return ok
		}
	})
}

// flushBuffer writes the buffer to w, updates remainingLimit, and reports whether streaming should continue.
func (s *StreamProxy[T]) flushBuffer(w io.Writer, buffer *LogsBuffer, remainingLimit *int) bool {
	written := buffer.Size()
	if err := buffer.Write(w); err != nil {
		log.Errorf("failed to write buffer: %v", err)
		return false
	}
	*remainingLimit -= written
	return !s.limitReached(*remainingLimit)
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
func (s *StreamProxy[T]) shouldFlushBuffer(buffer *LogsBuffer, remainingLimit int) bool {
	return buffer.Size() >= s.aggregator.GetBufferSize() || (remainingLimit > 0 && buffer.Size() >= remainingLimit)
}

// limitReached checks if the output limit has been reached
func (s *StreamProxy[T]) limitReached(remainingLimit int) bool {
	return remainingLimit <= 0 && s.limit > 0
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
