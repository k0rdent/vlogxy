package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsFieldMin struct {
	srcField string

	fieldName string
}

func (sm *statsFieldMin) Name() string {
	return "field_min"
}

func (sm *statsFieldMin) String() string {
	s := sm.Name() + "(" + quoteTokenIfNeeded(sm.srcField) + ", " + quoteTokenIfNeeded(sm.fieldName) + ")"
	return s
}

func (sm *statsFieldMin) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(sm.fieldName)
	pf.AddAllowFilter(sm.srcField)
}

func parseStatsFieldMin(lex *lexer) (statsFunc, error) {
	args, err := parseStatsFuncArgs(lex, "field_min")
	if err != nil {
		return nil, err
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("unexpected number of arguments for 'field_min' func; got %d args; want 2; args=%q", len(args), args)
	}

	sm := &statsFieldMin{
		srcField:  args[0],
		fieldName: args[1],
	}
	return sm, nil
}
