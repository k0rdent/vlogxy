package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsStddev struct {
	fieldFilters []string
}

func (ss *statsStddev) Name() string {
	return "stddev"
}

func (ss *statsStddev) String() string {
	return ss.Name() + "(" + fieldNamesString(ss.fieldFilters) + ")"
}

func (ss *statsStddev) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(ss.fieldFilters)
}

func parseStatsStddev(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "stddev")
	if err != nil {
		return nil, err
	}
	sa := &statsStddev{
		fieldFilters: fieldFilters,
	}
	return sa, nil
}
