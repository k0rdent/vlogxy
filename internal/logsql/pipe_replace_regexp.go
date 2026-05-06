package logstorage

import (
	"fmt"
	"regexp"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeReplaceRegexp processes '| replace_regexp ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#replace_regexp-pipe
type pipeReplaceRegexp struct {
	field string

	// re is the compiled regular expression
	re *regexp.Regexp

	// reStr contains string representation for the re.
	reStr string

	// replacement is the replacement string for the matching re.
	replacement string

	// limit limits the number of replacements, which can be performed
	limit uint64

	// iff is an optional filter for skipping the replace_regexp operation
	iff *ifFilter
}

func (pr *pipeReplaceRegexp) String() string {
	s := "replace_regexp"
	if pr.iff != nil {
		s += " " + pr.iff.String()
	}
	s += fmt.Sprintf(" (%s, %s)", quoteTokenIfNeeded(pr.reStr), quoteTokenIfNeeded(pr.replacement))
	if pr.field != "_msg" {
		s += " at " + quoteTokenIfNeeded(pr.field)
	}
	if pr.limit > 0 {
		s += fmt.Sprintf(" limit %d", pr.limit)
	}
	return s
}

func (pr *pipeReplaceRegexp) Name() string {
	return "replace_regexp"
}

func (pr *pipeReplaceRegexp) updateNeededFields(pf *prefixfilter.Filter) {
	updateNeededFieldsForUpdatePipe(pf, pr.field, pr.iff)
}

func (pr *pipeReplaceRegexp) visitSubqueries(visitFunc func(q *Query)) {
	pr.iff.visitSubqueries(visitFunc)
}

func parsePipeReplaceRegexp(lex *lexer) (pipe, error) {
	if !lex.isKeyword("replace_regexp") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "replace_regexp")
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
		return nil, fmt.Errorf("missing '(' after 'replace_regexp'")
	}
	lex.nextToken()

	reStr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot parse reStr in 'replace_regexp': %w", err)
	}
	re, err := regexpCompile(reStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse regexp %q in 'replace_regexp': %w", reStr, err)
	}
	if !lex.isKeyword(",") {
		return nil, fmt.Errorf("missing ',' after 'replace_regexp(%q'", reStr)
	}
	lex.nextToken()

	replacement, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot parse replacement in 'replace_regexp(%q': %w", reStr, err)
	}

	if !lex.isKeyword(")") {
		return nil, fmt.Errorf("missing ')' after 'replace_regexp(%q, %q'", reStr, replacement)
	}
	lex.nextToken()

	field := "_msg"
	if lex.isKeyword("at") {
		lex.nextToken()
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'at' field after 'replace_regexp(%q, %q)': %w", reStr, replacement, err)
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

	pr := &pipeReplaceRegexp{
		field:       field,
		re:          re,
		reStr:       reStr,
		replacement: replacement,
		limit:       limit,
		iff:         iff,
	}

	return pr, nil
}
