package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeSplit processes '| split ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#split-pipe
type pipeSplit struct {
	// separator is the separator for splitting the input field
	separator string

	// field to split
	srcField string

	// field to put the split result
	dstField string
}

func (ps *pipeSplit) String() string {
	s := "split " + quoteTokenIfNeeded(ps.separator)
	if ps.srcField != "_msg" {
		s += " from " + quoteTokenIfNeeded(ps.srcField)
	}
	if ps.dstField != ps.srcField {
		s += " as " + quoteTokenIfNeeded(ps.dstField)
	}
	return s
}

func (ps *pipeSplit) Name() string {
	return "split"
}

func (ps *pipeSplit) visitSubqueries(_ func(q *Query)) {
	// do nothing
}

func (ps *pipeSplit) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchString(ps.dstField) {
		pf.AddDenyFilter(ps.dstField)
		pf.AddAllowFilter(ps.srcField)
	}
}

func parsePipeSplit(lex *lexer) (pipe, error) {
	if !lex.isKeyword("split") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "split")
	}
	lex.nextToken()

	if lex.isKeyword("as", "from") {
		return nil, fmt.Errorf("missing split separator in front of %q", lex.token)
	}

	separator, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot read split separator: %w", err)
	}

	srcField := "_msg"
	if !lex.isKeyword("as", ")", "|", "") {
		if lex.isKeyword("from") {
			lex.nextToken()
		}
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse srcField name: %w", err)
		}
		srcField = field
	}

	dstField := srcField
	if !lex.isKeyword(")", "|", "") {
		if lex.isKeyword("as") {
			lex.nextToken()
		}
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse dstField name: %w", err)
		}
		dstField = field
	}

	ps := &pipeSplit{
		separator: separator,
		srcField:  srcField,
		dstField:  dstField,
	}

	return ps, nil
}
