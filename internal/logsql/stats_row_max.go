package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsRowMax struct {
	srcField string

	fieldFilters []string
}

func (sm *statsRowMax) Name() string {
	return "row_max"
}

func (sm *statsRowMax) String() string {
	s := sm.Name() + "(" + quoteTokenIfNeeded(sm.srcField)
	if !prefixfilter.MatchAll(sm.fieldFilters) {
		s += ", " + fieldNamesString(sm.fieldFilters)
	}
	s += ")"
	return s
}

func (sm *statsRowMax) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
	pf.AddAllowFilter(sm.srcField)
}

func parseStatsRowMax(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "row_max")
	if err != nil {
		return nil, err
	}

	if len(fieldFilters) == 0 {
		return nil, fmt.Errorf("missing source field for 'row_max' func")
	}

	srcField := fieldFilters[0]
	if prefixfilter.IsWildcardFilter(srcField) {
		return nil, fmt.Errorf("the source field %q cannot be wildcard", srcField)
	}

	fieldFilters = fieldFilters[1:]
	if len(fieldFilters) == 0 {
		fieldFilters = []string{"*"}
	}

	sm := &statsRowMax{
		srcField:     srcField,
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
