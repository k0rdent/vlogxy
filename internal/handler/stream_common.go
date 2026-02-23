package handler

import (
	"bufio"
	"bytes"
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// streamConfig holds common configuration shared by all stream aggregators.
type streamConfig struct {
	bufferSize int
	maxLimit   int
}

// baseStreamAggregator provides the common streaming behavior shared between
// StreamQuery and Tail. Embed this type to avoid duplicating the implementation.
type baseStreamAggregator struct {
	streamConfig
}

func newBaseStreamAggregator(maxLimit, bufferSize int) baseStreamAggregator {
	return baseStreamAggregator{
		streamConfig: streamConfig{
			bufferSize: bufferSize,
			maxLimit:   maxLimit,
		},
	}
}

// StreamParseResponse reads newline-delimited data from resp.Body and forwards
// each line to dataChan, stopping early when ctx is cancelled.
func (b *baseStreamAggregator) StreamParseResponse(ctx context.Context, resp *http.Response, dataChan chan<- []byte) {
	scanner := bufio.NewScanner(resp.Body)

	defer func() {
		if err := scanner.Err(); err != nil {
			log.Errorf("error reading response: %v", err)
		}
	}()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case dataChan <- bytes.Clone(scanner.Bytes()):
		}
	}
}

func (b *baseStreamAggregator) GetBufferSize() int {
	return b.bufferSize
}

func (b *baseStreamAggregator) GetMaxLogsLimit() int {
	return b.maxLimit
}
