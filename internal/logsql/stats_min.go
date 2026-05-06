package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsMin struct {
	fieldFilters []string
}

func (sm *statsMin) Name() string {
	return "min"
}

func (sm *statsMin) String() string {
	return sm.Name() + "(" + fieldNamesString(sm.fieldFilters) + ")"
}

func (sm *statsMin) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
}

func parseStatsMin(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "min")
	if err != nil {
		return nil, err
	}
	sm := &statsMin{
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
