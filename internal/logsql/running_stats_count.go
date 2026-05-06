package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type runningStatsCount struct {
	fieldFilters []string
}

func (sc *runningStatsCount) String() string {
	return "count(" + fieldNamesString(sc.fieldFilters) + ")"
}

func (sc *runningStatsCount) updateNeededFields(pf *prefixfilter.Filter) {
	if prefixfilter.MatchAll(sc.fieldFilters) {
		// Special case for count() - it doesn't need loading any additional fields
		return
	}

	pf.AddAllowFilters(sc.fieldFilters)
}

func parseRunningStatsCount(lex *lexer) (runningStatsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "count")
	if err != nil {
		return nil, err
	}
	sc := &runningStatsCount{
		fieldFilters: fieldFilters,
	}
	return sc, nil
}
