package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsUniqValues struct {
	fieldFilters []string
	limit        uint64
}

func (su *statsUniqValues) Name() string {
	return "uniq_values"
}

func (su *statsUniqValues) String() string {
	s := su.Name() + "(" + fieldNamesString(su.fieldFilters) + ")"
	if su.limit > 0 {
		s += fmt.Sprintf(" limit %d", su.limit)
	}
	return s
}

func (su *statsUniqValues) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(su.fieldFilters)
}

func parseStatsUniqValues(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "uniq_values")
	if err != nil {
		return nil, err
	}
	su := &statsUniqValues{
		fieldFilters: fieldFilters,
	}
	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		su.limit = n
	}
	return su, nil
}
