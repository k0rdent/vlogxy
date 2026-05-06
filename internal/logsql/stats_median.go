package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsMedian struct {
	sq *statsQuantile
}

func (sm *statsMedian) Name() string {
	return "median"
}

func (sm *statsMedian) String() string {
	return sm.Name() + "(" + fieldNamesString(sm.sq.fieldFilters) + ")"
}

func (sm *statsMedian) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.sq.fieldFilters)
}

func parseStatsMedian(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "median")
	if err != nil {
		return nil, err
	}
	sm := &statsMedian{
		sq: &statsQuantile{
			fieldFilters: fieldFilters,
			phi:          0.5,
			phiStr:       "0.5",
		},
	}
	return sm, nil
}
