package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type runningStatsSum struct {
	fieldFilters []string
}

func (ss *runningStatsSum) String() string {
	return "sum(" + fieldNamesString(ss.fieldFilters) + ")"
}

func (ss *runningStatsSum) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(ss.fieldFilters)
}

func parseRunningStatsSum(lex *lexer) (runningStatsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "sum")
	if err != nil {
		return nil, err
	}
	ss := &runningStatsSum{
		fieldFilters: fieldFilters,
	}
	return ss, nil
}
