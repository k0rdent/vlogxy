package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeExtract processes '| extract ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#extract-pipe
type pipeExtract struct {
	fromField string

	ptn        *pattern
	patternStr string

	keepOriginalFields bool
	skipEmptyResults   bool

	// iff is an optional filter for skipping the extract func
	iff *ifFilter
}

func (pe *pipeExtract) String() string {
	s := "extract"
	if pe.iff != nil {
		s += " " + pe.iff.String()
	}
	s += " " + quoteTokenIfNeeded(pe.patternStr)
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

func (pe *pipeExtract) Name() string {
	return "extract"
}

func (pe *pipeExtract) visitSubqueries(visitFunc func(q *Query)) {
	pe.iff.visitSubqueries(visitFunc)
}

func (pe *pipeExtract) updateNeededFields(pf *prefixfilter.Filter) {
	pfOrig := pf.Clone()
	needFromField := false
	for _, step := range pe.ptn.steps {
		if step.field == "" {
			continue
		}
		if pfOrig.MatchString(step.field) {
			needFromField = true
			if shouldDenyOverwrittenField(pe.iff, pe.keepOriginalFields, pe.skipEmptyResults) {
				pf.AddDenyFilter(step.field)
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

func parsePipeExtract(lex *lexer) (pipe, error) {
	if !lex.isKeyword("extract") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "extract")
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
	ptn, err := parsePattern(patternStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'pattern' %q: %w", patternStr, err)
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

	pe := &pipeExtract{
		fromField:          fromField,
		ptn:                ptn,
		patternStr:         patternStr,
		keepOriginalFields: keepOriginalFields,
		skipEmptyResults:   skipEmptyResults,
		iff:                iff,
	}

	return pe, nil
}
