package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsFieldMax struct {
	srcField string

	fieldName string
}

func (sm *statsFieldMax) Name() string {
	return "field_max"
}

func (sm *statsFieldMax) String() string {
	s := sm.Name() + "(" + quoteTokenIfNeeded(sm.srcField) + ", " + quoteTokenIfNeeded(sm.fieldName) + ")"
	return s
}

func (sm *statsFieldMax) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(sm.fieldName)
	pf.AddAllowFilter(sm.srcField)
}

func parseStatsFieldMax(lex *lexer) (statsFunc, error) {
	args, err := parseStatsFuncArgs(lex, "field_max")
	if err != nil {
		return nil, err
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("unexpected number of arguments for 'field_max' func; got %d args; want 2; args=%q", len(args), args)
	}

	sm := &statsFieldMax{
		srcField:  args[0],
		fieldName: args[1],
	}
	return sm, nil
}
