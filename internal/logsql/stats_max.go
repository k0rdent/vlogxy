package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsMax struct {
	fieldFilters []string
}

func (sm *statsMax) Name() string {
	return "max"
}

func (sm *statsMax) String() string {
	return sm.Name() + "(" + fieldNamesString(sm.fieldFilters) + ")"
}

func (sm *statsMax) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
}

func parseStatsMax(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "max")
	if err != nil {
		return nil, err
	}
	sm := &statsMax{
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
