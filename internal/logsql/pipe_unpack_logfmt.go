package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUnpackLogfmt processes '| unpack_logfmt ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#unpack_logfmt-pipe
type pipeUnpackLogfmt struct {
	// fromField is the field to unpack logfmt fields from
	fromField string

	// filterFields is list of field filters to extract from logfmt.
	fieldFilters []string

	// resultPrefix is prefix to add to unpacked field names
	resultPrefix string

	keepOriginalFields bool
	skipEmptyResults   bool

	// iff is an optional filter for skipping unpacking logfmt
	iff *ifFilter
}

func (pu *pipeUnpackLogfmt) String() string {
	s := "unpack_logfmt"
	if pu.iff != nil {
		s += " " + pu.iff.String()
	}
	if !isMsgFieldName(pu.fromField) {
		s += " from " + quoteTokenIfNeeded(pu.fromField)
	}
	if !prefixfilter.MatchAll(pu.fieldFilters) {
		s += " fields (" + fieldNamesString(pu.fieldFilters) + ")"
	}
	if pu.resultPrefix != "" {
		s += " result_prefix " + quoteTokenIfNeeded(pu.resultPrefix)
	}
	if pu.keepOriginalFields {
		s += " keep_original_fields"
	}
	if pu.skipEmptyResults {
		s += " skip_empty_results"
	}
	return s
}

func (pu *pipeUnpackLogfmt) Name() string {
	return "unpack_logfmt"
}

func (pu *pipeUnpackLogfmt) updateNeededFields(pf *prefixfilter.Filter) {
	updateNeededFieldsForUnpackPipe(pu.fromField, pu.resultPrefix, pu.fieldFilters, pu.keepOriginalFields, pu.skipEmptyResults, pu.iff, pf)
}

func (pu *pipeUnpackLogfmt) visitSubqueries(visitFunc func(q *Query)) {
	pu.iff.visitSubqueries(visitFunc)
}

func parsePipeUnpackLogfmt(lex *lexer) (pipe, error) {
	if !lex.isKeyword("unpack_logfmt") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "unpack_logfmt")
	}
	lex.nextToken()

	var iff *ifFilter
	if lex.isKeyword("if") {
		f, err := parseIfFilter(lex)
		if err != nil {
			return nil, err
		}
		iff = f
	}

	fromField := "_msg"
	if !lex.isKeyword("fields", "result_prefix", "keep_original_fields", "skip_empty_results", ")", "|", "") {
		if lex.isKeyword("from") {
			lex.nextToken()
		}
		f, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'from' field name: %w", err)
		}
		fromField = f
	}

	var fieldFilters []string
	if lex.isKeyword("fields") {
		lex.nextToken()
		fs, err := parseFieldFiltersInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'fields': %w", err)
		}
		fieldFilters = fs
	}
	if len(fieldFilters) == 0 {
		fieldFilters = []string{"*"}
	}

	resultPrefix := ""
	if lex.isKeyword("result_prefix") {
		lex.nextToken()
		p, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'result_prefix': %w", err)
		}
		resultPrefix = p
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

	pu := &pipeUnpackLogfmt{
		fromField:          fromField,
		fieldFilters:       fieldFilters,
		resultPrefix:       resultPrefix,
		keepOriginalFields: keepOriginalFields,
		skipEmptyResults:   skipEmptyResults,
		iff:                iff,
	}

	return pu, nil
}
