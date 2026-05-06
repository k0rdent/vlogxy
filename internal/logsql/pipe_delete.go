package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeDelete implements '| delete ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#delete-pipe
type pipeDelete struct {
	// fieldFilters contains a list of field filters to delete
	fieldFilters []string
}

func (pd *pipeDelete) String() string {
	if len(pd.fieldFilters) == 0 {
		logger.Panicf("BUG: pipeDelete must contain at least a single field")
	}

	return "delete " + fieldNamesString(pd.fieldFilters)
}

func (pd *pipeDelete) Name() string {
	return "delete"
}

func (pd *pipeDelete) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddDenyFilters(pd.fieldFilters)
}

func (pd *pipeDelete) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeDelete(lex *lexer) (pipe, error) {
	if !lex.isKeyword("delete", "del", "rm", "drop") {
		return nil, fmt.Errorf("expecting 'delete', 'del', 'rm' or 'drop'; got %q", lex.token)
	}
	lex.nextToken()

	fieldFilters, err := parseCommaSeparatedFields(lex)
	if err != nil {
		return nil, err
	}
	pd := &pipeDelete{
		fieldFilters: fieldFilters,
	}
	return pd, nil
}
