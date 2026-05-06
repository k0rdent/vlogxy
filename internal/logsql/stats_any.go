package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsAny struct {
	fieldName string
}

func (sa *statsAny) Name() string {
	return "any"
}

func (sa *statsAny) String() string {
	return sa.Name() + "(" + quoteTokenIfNeeded(sa.fieldName) + ")"
}

func (sa *statsAny) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(sa.fieldName)
}

func parseStatsAny(lex *lexer) (statsFunc, error) {
	args, err := parseStatsFuncArgs(lex, "any")
	if err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("unexpected number of args for 'any' function; got %d; want 1; args: %q", len(args), args)
	}

	sa := &statsAny{
		fieldName: args[0],
	}
	return sa, nil
}
