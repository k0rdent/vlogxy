package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsCount struct {
	fieldFilters []string
}

func (sc *statsCount) Name() string {
	return "count"
}

func (sc *statsCount) String() string {
	return sc.Name() + "(" + fieldNamesString(sc.fieldFilters) + ")"
}

func (sc *statsCount) updateNeededFields(pf *prefixfilter.Filter) {
	if prefixfilter.MatchAll(sc.fieldFilters) {
		// Special case for count() - it doesn't need loading any additional fields
		return
	}

	pf.AddAllowFilters(sc.fieldFilters)
}

func parseStatsCount(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "count")
	if err != nil {
		return nil, err
	}
	sc := &statsCount{
		fieldFilters: fieldFilters,
	}
	return sc, nil
}
