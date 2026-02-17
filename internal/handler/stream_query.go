package handler

import (
	"bufio"
	"bytes"
	"context"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	log "github.com/sirupsen/logrus"
)

// StreamQuery handles streaming of log query responses from multiple backends
type StreamQuery struct{}

// NewStreamQuery creates a new StreamQuery aggregator instance
func NewStreamQuery() interfaces.StreamResponseAggregator[[]byte] {
	return &StreamQuery{}
}

func (s *StreamQuery) StreamParseResponse(ctx context.Context, resp *http.Response, dataChan chan<- []byte) {
	scanner := bufio.NewScanner(resp.Body)

	defer func() {
		if err := scanner.Err(); err != nil {
			log.Errorf("error reading response: %v", err)
		}
	}()

	for scanner.Scan() {
		data := bytes.Clone(scanner.Bytes())

		select {
		case <-ctx.Done():
			return
		case dataChan <- data:
		}
	}
}
