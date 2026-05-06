package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUnpackJSON processes '| unpack_json ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#unpack_json-pipe
type pipeUnpackJSON struct {
	// fromField is the field to unpack json fields from
	fromField string

	// fieldFilters is a list of field filters to extract from json.
	fieldFilters []string

	// preserveKeys is a list of JSON keys for preserving JSON values.
	preserveKeys []string

	// resultPrefix is prefix to add to unpacked field names
	resultPrefix string

	keepOriginalFields bool
	skipEmptyResults   bool

	// iff is an optional filter for skipping unpacking json
	iff *ifFilter
}

func (pu *pipeUnpackJSON) String() string {
	s := "unpack_json"
	if pu.iff != nil {
		s += " " + pu.iff.String()
	}
	if !isMsgFieldName(pu.fromField) {
		s += " from " + quoteTokenIfNeeded(pu.fromField)
	}
	if !prefixfilter.MatchAll(pu.fieldFilters) {
		s += " fields (" + fieldNamesString(pu.fieldFilters) + ")"
	}
	if len(pu.preserveKeys) > 0 {
		s += " preserve_keys (" + fieldNamesString(pu.preserveKeys) + ")"
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

func (pu *pipeUnpackJSON) Name() string {
	return "unpack_json"
}

func (pu *pipeUnpackJSON) updateNeededFields(pf *prefixfilter.Filter) {
	updateNeededFieldsForUnpackPipe(pu.fromField, pu.resultPrefix, pu.fieldFilters, pu.keepOriginalFields, pu.skipEmptyResults, pu.iff, pf)
}

func (pu *pipeUnpackJSON) visitSubqueries(visitFunc func(q *Query)) {
	pu.iff.visitSubqueries(visitFunc)
}

func parsePipeUnpackJSON(lex *lexer) (pipe, error) {
	if !lex.isKeyword("unpack_json") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "unpack_json")
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
	if !lex.isKeyword("fields", "preserve_keys", "result_prefix", "keep_original_fields", "skip_empty_results", ")", "|", "") {
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

	var preserveKeys []string
	if lex.isKeyword("preserve_keys") {
		lex.nextToken()
		fn, err := parseFieldNamesInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'preserve_keys': %w", err)
		}
		preserveKeys = fn
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

	pu := &pipeUnpackJSON{
		fromField:          fromField,
		fieldFilters:       fieldFilters,
		preserveKeys:       preserveKeys,
		resultPrefix:       resultPrefix,
		keepOriginalFields: keepOriginalFields,
		skipEmptyResults:   skipEmptyResults,
		iff:                iff,
	}

	return pu, nil
}
