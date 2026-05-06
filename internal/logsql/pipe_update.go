package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

func updateNeededFieldsForUpdatePipe(pf *prefixfilter.Filter, field string, iff *ifFilter) {
	if iff != nil && pf.MatchString(field) {
		pf.AddAllowFilters(iff.allowFilters)
	}
}

// shouldDenyOverwrittenField reports whether planner can safely avoid reading
// the original value for an overwritten field.
func shouldDenyOverwrittenField(iff *ifFilter, keepOriginalFields, skipEmptyResults bool) bool {
	return iff == nil && !keepOriginalFields && !skipEmptyResults
}
