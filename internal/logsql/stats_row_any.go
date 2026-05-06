package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsRowAny struct {
	fieldFilters []string
}

func (sa *statsRowAny) Name() string {
	return "row_any"
}

func (sa *statsRowAny) String() string {
	return sa.Name() + "(" + fieldNamesString(sa.fieldFilters) + ")"
}

func (sa *statsRowAny) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sa.fieldFilters)
}

func parseStatsRowAny(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "row_any")
	if err != nil {
		return nil, err
	}

	sa := &statsRowAny{
		fieldFilters: fieldFilters,
	}
	return sa, nil
}
