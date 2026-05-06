package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeCollapseNums processes '| collapse_nums ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#collapse_nums-pipe
type pipeCollapseNums struct {
	// the field to collapse nums at
	field string

	// if isPrettify is set, then collapsed nums are prettified with common placeholders
	isPrettify bool

	// iff is an optional filter for skipping the collapse_nums operation
	iff *ifFilter
}

func (pc *pipeCollapseNums) String() string {
	s := "collapse_nums"
	if pc.iff != nil {
		s += " " + pc.iff.String()
	}
	if pc.field != "_msg" {
		s += " at " + quoteTokenIfNeeded(pc.field)
	}
	if pc.isPrettify {
		s += " prettify"
	}
	return s
}

func (pc *pipeCollapseNums) Name() string {
	return "collapse_nums"
}

func (pc *pipeCollapseNums) updateNeededFields(pf *prefixfilter.Filter) {
	updateNeededFieldsForUpdatePipe(pf, pc.field, pc.iff)
}

func (pc *pipeCollapseNums) visitSubqueries(visitFunc func(q *Query)) {
	pc.iff.visitSubqueries(visitFunc)
}

func parsePipeCollapseNums(lex *lexer) (pipe, error) {
	if !lex.isKeyword("collapse_nums") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "collapse_nums")
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

	field := "_msg"
	if lex.isKeyword("at") {
		lex.nextToken()
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'at' field after 'collapse_nums': %w", err)
		}
		field = f
	}

	isPrettify := false
	if lex.isKeyword("prettify") {
		lex.nextToken()
		isPrettify = true
	}

	pc := &pipeCollapseNums{
		field:      field,
		isPrettify: isPrettify,
		iff:        iff,
	}

	return pc, nil
}
