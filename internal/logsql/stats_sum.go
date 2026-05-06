package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsSum struct {
	fieldFilters []string
}

func (ss *statsSum) Name() string {
	return "sum"
}

func (ss *statsSum) String() string {
	return ss.Name() + "(" + fieldNamesString(ss.fieldFilters) + ")"
}

func (ss *statsSum) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(ss.fieldFilters)
}

func parseStatsSum(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "sum")
	if err != nil {
		return nil, err
	}
	ss := &statsSum{
		fieldFilters: fieldFilters,
	}
	return ss, nil
}
