package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type runningStatsMin struct {
	fieldFilters []string
}

func (sm *runningStatsMin) String() string {
	return "min(" + fieldNamesString(sm.fieldFilters) + ")"
}

func (sm *runningStatsMin) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
}

func parseRunningStatsMin(lex *lexer) (runningStatsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "min")
	if err != nil {
		return nil, err
	}
	sm := &runningStatsMin{
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
