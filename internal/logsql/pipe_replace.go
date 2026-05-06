package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeReplace processes '| replace ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#replace-pipe
type pipeReplace struct {
	field     string
	oldSubstr string
	newSubstr string

	// limit limits the number of replacements, which can be performed
	limit uint64

	// iff is an optional filter for skipping the replace operation
	iff *ifFilter
}

func (pr *pipeReplace) String() string {
	s := "replace"
	if pr.iff != nil {
		s += " " + pr.iff.String()
	}
	s += fmt.Sprintf(" (%s, %s)", quoteTokenIfNeeded(pr.oldSubstr), quoteTokenIfNeeded(pr.newSubstr))
	if pr.field != "_msg" {
		s += " at " + quoteTokenIfNeeded(pr.field)
	}
	if pr.limit > 0 {
		s += fmt.Sprintf(" limit %d", pr.limit)
	}
	return s
}

func (pr *pipeReplace) Name() string {
	return "replace"
}

func (pr *pipeReplace) updateNeededFields(pf *prefixfilter.Filter) {
	updateNeededFieldsForUpdatePipe(pf, pr.field, pr.iff)
}

func (pr *pipeReplace) visitSubqueries(visitFunc func(q *Query)) {
	pr.iff.visitSubqueries(visitFunc)
}

func parsePipeReplace(lex *lexer) (pipe, error) {
	if !lex.isKeyword("replace") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "replace")
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

	if !lex.isKeyword("(") {
		return nil, fmt.Errorf("missing '(' after 'replace'")
	}
	lex.nextToken()

	oldSubstr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot parse oldSubstr in 'replace': %w", err)
	}
	if !lex.isKeyword(",") {
		return nil, fmt.Errorf("missing ',' after 'replace(%q'", oldSubstr)
	}
	lex.nextToken()

	newSubstr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot parse newSubstr in 'replace(%q': %w", oldSubstr, err)
	}

	if !lex.isKeyword(")") {
		return nil, fmt.Errorf("missing ')' after 'replace(%q, %q'", oldSubstr, newSubstr)
	}
	lex.nextToken()

	field := "_msg"
	if lex.isKeyword("at") {
		lex.nextToken()
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'at' field after 'replace(%q, %q)': %w", oldSubstr, newSubstr, err)
		}
		field = f
	}

	limit := uint64(0)
	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		limit = n
	}

	pr := &pipeReplace{
		field:     field,
		oldSubstr: oldSubstr,
		newSubstr: newSubstr,
		limit:     limit,
		iff:       iff,
	}

	return pr, nil
}
