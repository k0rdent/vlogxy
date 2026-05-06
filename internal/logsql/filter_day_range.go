package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterDayRange filters by day range.
//
// It is expressed as `_time:day_range[start, end] offset d` in LogsQL.
type filterDayRange struct {
	// start is the offset in nanoseconds from the beginning of the day for the day range start.
	start int64

	// end is the offset in nanoseconds from the beginning of the day for the day range end.
	end int64

	// offset is the offset, which must be applied to _time before applying [start, end] filter to it.
	offset int64

	// stringRepr is string representation of the filter.
	stringRepr string
}

func newFilterDayRange(start, end, offset int64, stringRepr string) *filterDayRange {
	return &filterDayRange{
		start:      start,
		end:        end,
		offset:     offset,
		stringRepr: stringRepr,
	}
}

func (fr *filterDayRange) String() string {
	return "_time:day_range" + fr.stringRepr
}

func (fr *filterDayRange) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_time")
}
