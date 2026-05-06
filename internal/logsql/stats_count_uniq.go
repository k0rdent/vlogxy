package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsCountUniq struct {
	fields []string
	limit  uint64
}

func (su *statsCountUniq) Name() string {
	return "count_uniq"
}

func (su *statsCountUniq) String() string {
	s := su.Name() + "(" + fieldNamesString(su.fields) + ")"
	if su.limit > 0 {
		s += fmt.Sprintf(" limit %d", su.limit)
	}
	return s
}

func (su *statsCountUniq) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(su.fields)
}

func parseStatsCountUniq(lex *lexer) (statsFunc, error) {
	fields, err := parseStatsFuncFields(lex, "count_uniq")
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("expecting at least a single field")
	}
	su := &statsCountUniq{
		fields: fields,
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
