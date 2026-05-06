package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type runningStatsMax struct {
	fieldFilters []string
}

func (sm *runningStatsMax) String() string {
	return "max(" + fieldNamesString(sm.fieldFilters) + ")"
}

func (sm *runningStatsMax) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
}

func parseRunningStatsMax(lex *lexer) (runningStatsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "max")
	if err != nil {
		return nil, err
	}
	sm := &runningStatsMax{
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
