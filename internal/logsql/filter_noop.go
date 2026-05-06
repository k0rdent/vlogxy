package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterNoop does nothing
type filterNoop struct {
}

func newFilterNoop() *filterNoop {
	return &noopFilter
}

var noopFilter filterNoop

func (fn *filterNoop) String() string {
	return "*"
}

func (fn *filterNoop) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}
