package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsRowMin struct {
	srcField string

	fieldFilters []string
}

func (sm *statsRowMin) Name() string {
	return "row_min"
}

func (sm *statsRowMin) String() string {
	s := sm.Name() + "(" + quoteTokenIfNeeded(sm.srcField)
	if !prefixfilter.MatchAll(sm.fieldFilters) {
		s += ", " + fieldNamesString(sm.fieldFilters)
	}
	s += ")"
	return s
}

func (sm *statsRowMin) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sm.fieldFilters)
	pf.AddAllowFilter(sm.srcField)
}

func parseStatsRowMin(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "row_min")
	if err != nil {
		return nil, err
	}

	if len(fieldFilters) == 0 {
		return nil, fmt.Errorf("missing source field for 'row_min' func")
	}

	srcField := fieldFilters[0]
	if prefixfilter.IsWildcardFilter(srcField) {
		return nil, fmt.Errorf("the source field %q cannot be wildcard", srcField)
	}

	fieldFilters = fieldFilters[1:]
	if len(fieldFilters) == 0 {
		fieldFilters = []string{"*"}
	}

	sm := &statsRowMin{
		srcField:     srcField,
		fieldFilters: fieldFilters,
	}
	return sm, nil
}
