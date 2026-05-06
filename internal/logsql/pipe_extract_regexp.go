package logstorage

import (
	"fmt"
	"regexp"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeExtractRegexp processes '| extract_regexp ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#extract_regexp-pipe
type pipeExtractRegexp struct {
	fromField string

	// re is compiled regular expression for the matching pattern
	re *regexp.Regexp

	// reStr is string representation for re.
	reStr string

	// reFields contains named capturing fields from the re.
	reFields []string

	keepOriginalFields bool
	skipEmptyResults   bool

	// iff is an optional filter for skipping the extract func
	iff *ifFilter
}

func (pe *pipeExtractRegexp) String() string {
	s := "extract_regexp"
	if pe.iff != nil {
		s += " " + pe.iff.String()
	}
	s += " " + quoteTokenIfNeeded(pe.reStr)
	if !isMsgFieldName(pe.fromField) {
		s += " from " + quoteTokenIfNeeded(pe.fromField)
	}
	if pe.keepOriginalFields {
		s += " keep_original_fields"
	}
	if pe.skipEmptyResults {
		s += " skip_empty_results"
	}
	return s
}

func (pe *pipeExtractRegexp) Name() string {
	return "extract_regexp"
}

func (pe *pipeExtractRegexp) visitSubqueries(visitFunc func(q *Query)) {
	pe.iff.visitSubqueries(visitFunc)
}

func (pe *pipeExtractRegexp) updateNeededFields(pf *prefixfilter.Filter) {
	pfOrig := pf.Clone()
	needFromField := false
	for _, f := range pe.reFields {
		if f == "" {
			continue
		}
		if pfOrig.MatchString(f) {
			needFromField = true
			if shouldDenyOverwrittenField(pe.iff, pe.keepOriginalFields, pe.skipEmptyResults) {
				pf.AddDenyFilter(f)
			}
		}
	}
	if needFromField {
		pf.AddAllowFilter(pe.fromField)
		if pe.iff != nil {
			pf.AddAllowFilters(pe.iff.allowFilters)
		}
	}
}

func parsePipeExtractRegexp(lex *lexer) (pipe, error) {
	if !lex.isKeyword("extract_regexp") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "extract_regexp")
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

	// parse pattern
	patternStr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot read 'pattern': %w", err)
	}

	re, err := regexpCompile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'pattern' %q: %w", patternStr, err)
	}
	reFields := re.SubexpNames()

	hasNamedFields := false
	for _, f := range reFields {
		if f != "" {
			hasNamedFields = true
			break
		}
	}
	if !hasNamedFields {
		return nil, fmt.Errorf("the 'pattern' %q must contain at least a single named group in the form (?P<group_name>...)", patternStr)
	}

	// parse optional 'from ...' part
	fromField := "_msg"
	if lex.isKeyword("from") {
		lex.nextToken()
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'from' field name: %w", err)
		}
		fromField = f
	}

	keepOriginalFields := false
	skipEmptyResults := false
	switch {
	case lex.isKeyword("keep_original_fields"):
		lex.nextToken()
		keepOriginalFields = true
	case lex.isKeyword("skip_empty_results"):
		lex.nextToken()
		skipEmptyResults = true
	}

	pe := &pipeExtractRegexp{
		fromField:          fromField,
		re:                 re,
		reStr:              patternStr,
		reFields:           reFields,
		keepOriginalFields: keepOriginalFields,
		skipEmptyResults:   skipEmptyResults,
		iff:                iff,
	}

	return pe, nil
}

func regexpCompile(s string) (*regexp.Regexp, error) {
	// Make sure that '.' inside the patternStr matches newline chars.
	// See https://github.com/VictoriaMetrics/VictoriaLogs/issues/88
	s = "(?s)(?:" + s + ")"

	return regexp.Compile(s)
}
