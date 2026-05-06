package logstorage

import (
	"fmt"
	"slices"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUnroll processes '| unroll ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#unroll-pipe
type pipeUnroll struct {
	// fields to unroll
	fields []string

	// iff is an optional filter for skipping the unroll
	iff *ifFilter
}

func (pu *pipeUnroll) String() string {
	s := "unroll"
	if pu.iff != nil {
		s += " " + pu.iff.String()
	}
	s += " by (" + fieldNamesString(pu.fields) + ")"
	return s
}

func (pu *pipeUnroll) Name() string {
	return "unroll"
}

func (pu *pipeUnroll) visitSubqueries(visitFunc func(q *Query)) {
	pu.iff.visitSubqueries(visitFunc)
}

func (pu *pipeUnroll) updateNeededFields(pf *prefixfilter.Filter) {
	if pu.iff != nil {
		pf.AddAllowFilters(pu.iff.allowFilters)
	}
	pf.AddAllowFilters(pu.fields)
}

func parsePipeUnroll(lex *lexer) (pipe, error) {
	if !lex.isKeyword("unroll") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "unroll")
	}
	lex.nextToken()

	// parse optional if (...)
	var iff *ifFilter
	if lex.isKeyword("if") {
		f, err := parseIfFilter(lex)
		if err != nil {
			return nil, err
		}
		iff = f
	}

	// parse by (...)
	if lex.isKeyword("by") {
		lex.nextToken()
	}

	var fields []string
	if lex.isKeyword("(") {
		fs, err := parseFieldNamesInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by(...)': %w", err)
		}
		fields = fs
	} else {
		fs, err := parseCommaSeparatedFields(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by ...': %w", err)
		}
		fields = fs
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("'by(...)' at 'unroll' must contain at least a single field")
	}
	if slices.Contains(fields, "*") {
		return nil, fmt.Errorf("unroll by '*' isn't supported")
	}

	pu := &pipeUnroll{
		fields: fields,
		iff:    iff,
	}

	return pu, nil
}
