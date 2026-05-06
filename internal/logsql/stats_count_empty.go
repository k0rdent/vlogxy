package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsCountEmpty struct {
	fieldFilters []string
}

func (sc *statsCountEmpty) Name() string {
	return "count_empty"
}

func (sc *statsCountEmpty) String() string {
	return sc.Name() + "(" + fieldNamesString(sc.fieldFilters) + ")"
}

func (sc *statsCountEmpty) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sc.fieldFilters)
}

func parseStatsCountEmpty(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "count_empty")
	if err != nil {
		return nil, err
	}
	sc := &statsCountEmpty{
		fieldFilters: fieldFilters,
	}
	return sc, nil
}
