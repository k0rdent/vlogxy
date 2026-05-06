package logstorage

import (
	"strings"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterAnd contains filters joined by AND operator.
//
// It is expressed as `f1 AND f2 ... AND fN` in LogsQL.
type filterAnd struct {
	filters []filter
}

func newFilterAnd(filters []filter) *filterAnd {
	return &filterAnd{
		filters: filters,
	}
}

func (fa *filterAnd) String() string {
	filters := fa.filters
	a := make([]string, len(filters))
	for i, f := range filters {
		s := f.String()
		if _, ok := f.(*filterOr); ok {
			s = "(" + s + ")"
		}
		a[i] = s
	}
	return strings.Join(a, " ")
}

func (fa *filterAnd) updateNeededFields(pf *prefixfilter.Filter) {
	for _, f := range fa.filters {
		f.updateNeededFields(pf)
	}
}
