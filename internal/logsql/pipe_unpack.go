package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

func updateNeededFieldsForUnpackPipe(fromField, outFieldPrefix string, outFieldFilters []string, keepOriginalFields, skipEmptyResults bool, iff *ifFilter, pf *prefixfilter.Filter) {
	if pf.MatchNothing() {
		// There is no need in fetching any fields, since the caller ignores all the fields.
		return
	}

	needFromField := len(outFieldFilters) == 0
	for _, f := range outFieldFilters {
		if pf.MatchStringOrWildcard(outFieldPrefix + f) {
			needFromField = true
			break
		}
	}
	if !keepOriginalFields && !skipEmptyResults {
		for _, f := range outFieldFilters {
			if !prefixfilter.IsWildcardFilter(f) {
				pf.AddDenyFilter(outFieldPrefix + f)
			}
		}
	}
	if needFromField {
		pf.AddAllowFilter(fromField)
		if iff != nil {
			pf.AddAllowFilters(iff.allowFilters)
		}
	}
}
