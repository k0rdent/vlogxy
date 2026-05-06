package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsValues struct {
	fieldFilters []string
	limit        uint64
}

func (sv *statsValues) Name() string {
	return "values"
}

func (sv *statsValues) String() string {
	s := sv.Name() + "(" + fieldNamesString(sv.fieldFilters) + ")"
	if sv.limit > 0 {
		s += fmt.Sprintf(" limit %d", sv.limit)
	}
	return s
}

func (sv *statsValues) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sv.fieldFilters)
}

func parseStatsValues(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "values")
	if err != nil {
		return nil, err
	}
	sv := &statsValues{
		fieldFilters: fieldFilters,
	}
	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		sv.limit = n
	}
	return sv, nil
}
