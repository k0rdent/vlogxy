package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFieldValues processes '| field_values ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#field_values-pipe
type pipeFieldValues struct {
	field string

	// if the filter is non-empty then only the field values containing the given filter substring are returned.
	filter string

	limit uint64
}

func (pf *pipeFieldValues) String() string {
	s := "field_values " + quoteTokenIfNeeded(pf.field)
	if pf.filter != "" {
		s += " filter " + quoteTokenIfNeeded(pf.filter)
	}
	if pf.limit > 0 {
		s += fmt.Sprintf(" limit %d", pf.limit)
	}
	return s
}

func (pf *pipeFieldValues) Name() string {
	return "field_values"
}

func (pf *pipeFieldValues) updateNeededFields(f *prefixfilter.Filter) {
	f.Reset()
	f.AddAllowFilter(pf.field)
}

func (pf *pipeFieldValues) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeFieldValues(lex *lexer) (pipe, error) {
	if !lex.isKeyword("field_values") {
		return nil, fmt.Errorf("expecting 'field_values'; got %q", lex.token)
	}
	lex.nextToken()

	field, err := parseFieldNameWithOptionalParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse field name for 'field_values': %w", err)
	}

	filter := ""
	if lex.isKeyword("filter") {
		lex.nextToken()
		f, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot parse filter for 'field_values': %w", err)
		}
		filter = f
	}

	limit := uint64(0)
	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		limit = n
	}

	pf := &pipeFieldValues{
		field:  field,
		filter: filter,
		limit:  limit,
	}

	return pf, nil
}
