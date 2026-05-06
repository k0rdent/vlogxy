package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsCountUniqHash struct {
	fields []string
	limit  uint64
}

func (su *statsCountUniqHash) Name() string {
	return "count_uniq_hash"
}

func (su *statsCountUniqHash) String() string {
	s := su.Name() + "(" + fieldNamesString(su.fields) + ")"
	if su.limit > 0 {
		s += fmt.Sprintf(" limit %d", su.limit)
	}
	return s
}

func (su *statsCountUniqHash) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(su.fields)
}

func parseStatsCountUniqHash(lex *lexer) (statsFunc, error) {
	fields, err := parseStatsFuncFields(lex, "count_uniq_hash")
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("expecting at least a single field")
	}
	su := &statsCountUniqHash{
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
