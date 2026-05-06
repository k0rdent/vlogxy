package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeDropEmptyFields processes '| drop_empty_fields ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#drop_empty_fields-pipe
type pipeDropEmptyFields struct {
}

func (pd *pipeDropEmptyFields) String() string {
	return "drop_empty_fields"
}

func (pd *pipeDropEmptyFields) Name() string {
	return pd.String()
}

func (pd *pipeDropEmptyFields) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func (pd *pipeDropEmptyFields) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func parsePipeDropEmptyFields(lex *lexer) (pipe, error) {
	if !lex.isKeyword("drop_empty_fields") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "drop_empty_fields")
	}
	lex.nextToken()

	pd := &pipeDropEmptyFields{}

	return pd, nil
}
