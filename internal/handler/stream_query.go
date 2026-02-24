package handler

import "github.com/k0rdent/vlogxy/internal/interfaces"

// StreamQuery aggregates streaming log query responses from multiple backends.
type StreamQuery struct {
	baseStreamAggregator
}

// NewStreamQuery creates a new StreamQuery aggregator instance.
func NewStreamQuery(maxLimit, bufferSize int) interfaces.StreamResponseAggregator[[]byte] {
	return &StreamQuery{baseStreamAggregator: newBaseStreamAggregator(maxLimit, bufferSize)}
}
