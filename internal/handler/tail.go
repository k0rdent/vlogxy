package handler

import "github.com/k0rdent/vlogxy/internal/interfaces"

// Tail aggregates streaming tail-log responses from multiple backends.
type Tail struct {
	baseStreamAggregator
}

// NewTail creates a new Tail aggregator.
// maxLimit=0 means no limit (appropriate for a continuous tail stream).
func NewTail(maxLimit, bufferSize int) interfaces.StreamResponseAggregator[[]byte] {
	return &Tail{baseStreamAggregator: newBaseStreamAggregator(maxLimit, bufferSize)}
}
