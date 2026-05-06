package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterTime filters by time.
//
// It is expressed as `_time:[start, end]` in LogsQL.
type filterTime struct {
	// mintimestamp is the minimum timestamp in nanoseconds to find
	minTimestamp int64

	// maxTimestamp is the maximum timestamp in nanoseconds to find
	maxTimestamp int64

	// stringRepr is string representation of the filter
	stringRepr string
}

func newFilterTime(minTimestamp, maxTimestamp int64, stringRepr string) *filterTime {
	return &filterTime{
		minTimestamp: minTimestamp,
		maxTimestamp: maxTimestamp,

		stringRepr: stringRepr,
	}
}

func (ft *filterTime) String() string {
	return "_time:" + ft.stringRepr
}

func (ft *filterTime) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_time")
}
