package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

func updateNeededFieldsForPipePack(pf *prefixfilter.Filter, resultField string, fieldFilters []string) {
	if pf.MatchString(resultField) {
		pf.AddDenyFilter(resultField)
		if len(fieldFilters) > 0 {
			pf.AddAllowFilters(fieldFilters)
		} else {
			pf.AddAllowFilter("*")
		}
	}
}
