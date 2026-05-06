package logstorage

import (
	"time"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterWeekRange filters by week range.
//
// It is expressed as `_time:week_range[start, end] offset d` in LogsQL.
type filterWeekRange struct {
	// startDay is the starting day of the week.
	startDay time.Weekday

	// endDay is the ending day of the week.
	endDay time.Weekday

	// offset is the offset, which must be applied to _time before applying [start, end] filter to it.
	offset int64

	// stringRepr is string representation of the filter.
	stringRepr string
}

func newFilterWeekRange(startDay, endDay time.Weekday, offset int64, stringRepr string) *filterWeekRange {
	return &filterWeekRange{
		startDay:   startDay,
		endDay:     endDay,
		offset:     offset,
		stringRepr: stringRepr,
	}
}

func (fr *filterWeekRange) String() string {
	return "_time:week_range" + fr.stringRepr
}

func (fr *filterWeekRange) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_time")
}
