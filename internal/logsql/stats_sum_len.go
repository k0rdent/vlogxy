package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsSumLen struct {
	fieldFilters []string
}

func (ss *statsSumLen) Name() string {
	return "sum_len"
}

func (ss *statsSumLen) String() string {
	return ss.Name() + "(" + fieldNamesString(ss.fieldFilters) + ")"
}

func (ss *statsSumLen) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(ss.fieldFilters)
}

func parseStatsSumLen(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "sum_len")
	if err != nil {
		return nil, err
	}
	ss := &statsSumLen{
		fieldFilters: fieldFilters,
	}
	return ss, nil
}
