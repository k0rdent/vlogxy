package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFormat processes '| format ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#format-pipe
type pipeFormat struct {
	formatStr string
	steps     []patternStep

	resultField string

	keepOriginalFields bool
	skipEmptyResults   bool

	// iff is an optional filter for skipping the format func
	iff *ifFilter
}

func (pf *pipeFormat) String() string {
	s := "format"
	if pf.iff != nil {
		s += " " + pf.iff.String()
	}
	s += " " + quoteTokenIfNeeded(pf.formatStr)
	if !isMsgFieldName(pf.resultField) {
		s += " as " + quoteTokenIfNeeded(pf.resultField)
	}
	if pf.keepOriginalFields {
		s += " keep_original_fields"
	}
	if pf.skipEmptyResults {
		s += " skip_empty_results"
	}
	return s
}

func (pf *pipeFormat) Name() string {
	return "format"
}

func (pf *pipeFormat) updateNeededFields(f *prefixfilter.Filter) {
	if !f.MatchString(pf.resultField) {
		return
	}

	if pf.iff != nil {
		f.AddAllowFilters(pf.iff.allowFilters)
	} else if shouldDenyOverwrittenField(pf.iff, pf.keepOriginalFields, pf.skipEmptyResults) {
		f.AddDenyFilter(pf.resultField)
	}
	for _, step := range pf.steps {
		if step.field != "" {
			f.AddAllowFilter(step.field)
		}
	}
}

func (pf *pipeFormat) visitSubqueries(visitFunc func(q *Query)) {
	pf.iff.visitSubqueries(visitFunc)
}

func parsePipeFormat(lex *lexer) (pipe, error) {
	if !lex.isKeyword("format") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "format")
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

	// parse format
	formatStr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot read 'format': %w", err)
	}
	steps, err := parsePatternSteps(formatStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'pattern' %q: %w", formatStr, err)
	}

	// Verify that all the fields mentioned in the format pattern do not end with '*'
	for _, step := range steps {
		if prefixfilter.IsWildcardFilter(step.field) {
			return nil, fmt.Errorf("'pattern' %q cannot contain wildcard fields like %q", formatStr, step.field)
		}
	}

	// parse optional 'as ...` part
	resultField := "_msg"
	if lex.isKeyword("as") {
		lex.nextToken()
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result field after 'format %q as': %w", formatStr, err)
		}
		resultField = field
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

	pf := &pipeFormat{
		formatStr:          formatStr,
		steps:              steps,
		resultField:        resultField,
		keepOriginalFields: keepOriginalFields,
		skipEmptyResults:   skipEmptyResults,
		iff:                iff,
	}

	return pf, nil
}
