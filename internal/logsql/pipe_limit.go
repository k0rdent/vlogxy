package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeLimit implements '| limit ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#limit-pipe
type pipeLimit struct {
	limit uint64
}

func (pl *pipeLimit) String() string {
	return fmt.Sprintf("limit %d", pl.limit)
}

func (pl *pipeLimit) Name() string {
	return "limit"
}

func (pl *pipeLimit) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func (pl *pipeLimit) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}
