package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeDecolorize processes '| decolorize ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#decolorize-pipe
type pipeDecolorize struct {
	field string
}

func (pd *pipeDecolorize) String() string {
	s := "decolorize"
	if pd.field != "_msg" {
		s += " " + quoteTokenIfNeeded(pd.field)
	}
	return s
}

func (pd *pipeDecolorize) Name() string {
	return "decolorize"
}

func (pd *pipeDecolorize) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func (pd *pipeDecolorize) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeDecolorize(lex *lexer) (pipe, error) {
	if !lex.isKeyword("decolorize") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "decolorize")
	}
	lex.nextToken()

	field := "_msg"
	if !lex.isKeyword("|", ")", "") {
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse field name after 'decolorize': %w", err)
		}
		field = f
	}

	pd := &pipeDecolorize{
		field: field,
	}

	return pd, nil
}
