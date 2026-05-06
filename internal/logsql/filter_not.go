package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterNot negates the filter.
//
// It is expressed as `NOT f` or `!f` in LogsQL.
type filterNot struct {
	f filter
}

func newFilterNot(f filter) *filterNot {
	return &filterNot{
		f: f,
	}
}

func (fn *filterNot) String() string {
	s := fn.f.String()
	switch fn.f.(type) {
	case *filterAnd, *filterOr:
		return "!(" + s + ")"
	default:
		return "!" + s
	}
}

func (fn *filterNot) updateNeededFields(pf *prefixfilter.Filter) {
	fn.f.updateNeededFields(pf)
}
