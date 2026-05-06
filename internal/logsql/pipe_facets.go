package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFacetsDefaultLimit is the default number of entries pipeFacets returns per each log field.
const pipeFacetsDefaultLimit = 10

// pipeFacetsDefaultMaxValuesPerField is the default number of unique values to track per each field.
const pipeFacetsDefaultMaxValuesPerField = 1000

// pipeFacetsDefaultMaxValueLen is the default length of values in fields, which must be ignored when building facets.
const pipeFacetsDefaultMaxValueLen = 128

// pipeFacets processes '| facets ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#facets-pipe
type pipeFacets struct {
	// limit is the maximum number of values to return per each field with the maximum number of hits.
	limit uint64

	// the maximum unique values to track per each field. Fields with bigger number of unique values are ignored.
	maxValuesPerField uint64

	// fields with values longer than maxValueLen are ignored, since it is hard to use them in faceted search.
	maxValueLen uint64

	// keep facets for fields with const values over all the selected logs.
	//
	// by default such fields are skipped, since they do not help investigating the selected logs.
	keepConstFields bool
}

func (pf *pipeFacets) String() string {
	s := "facets"
	if pf.limit != pipeFacetsDefaultLimit {
		s += fmt.Sprintf(" %d", pf.limit)
	}
	if pf.maxValuesPerField != pipeFacetsDefaultMaxValuesPerField {
		s += fmt.Sprintf(" max_values_per_field %d", pf.maxValuesPerField)
	}
	if pf.maxValueLen != pipeFacetsDefaultMaxValueLen {
		s += fmt.Sprintf(" max_value_len %d", pf.maxValueLen)
	}
	if pf.keepConstFields {
		s += " keep_const_fields"
	}
	return s
}

func (pf *pipeFacets) Name() string {
	return "facets"
}

func (pf *pipeFacets) updateNeededFields(f *prefixfilter.Filter) {
	f.AddAllowFilter("*")
}

func (pf *pipeFacets) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeFacets(lex *lexer) (pipe, error) {
	if !lex.isKeyword("facets") {
		return nil, fmt.Errorf("expecting 'facets'; got %q", lex.token)
	}
	lex.nextToken()

	limit := uint64(pipeFacetsDefaultLimit)
	if isNumberPrefix(lex.token) {
		limitF, s, err := parseNumber(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse N in 'facets': %w", err)
		}
		if limitF < 1 {
			return nil, fmt.Errorf("value N in 'facets %s' must be integer bigger than 0", s)
		}
		limit = uint64(limitF)
	}

	pf := &pipeFacets{
		limit:             limit,
		maxValuesPerField: pipeFacetsDefaultMaxValuesPerField,
		maxValueLen:       pipeFacetsDefaultMaxValueLen,
	}
	for {
		switch {
		case lex.isKeyword("max_values_per_field"):
			lex.nextToken()
			n, s, err := parseNumber(lex)
			if err != nil {
				return nil, fmt.Errorf("cannot parse max_values_per_field: %w", err)
			}
			if n < 1 {
				return nil, fmt.Errorf("max_value_per_field must be integer bigger than 0; got %s", s)
			}
			pf.maxValuesPerField = uint64(n)
		case lex.isKeyword("max_value_len"):
			lex.nextToken()
			n, s, err := parseNumber(lex)
			if err != nil {
				return nil, fmt.Errorf("cannot parse max_value_len: %w", err)
			}
			if n < 1 {
				return nil, fmt.Errorf("max_value_len must be integer bigger than 0; got %s", s)
			}
			pf.maxValueLen = uint64(n)
		case lex.isKeyword("keep_const_fields"):
			lex.nextToken()
			pf.keepConstFields = true
		default:
			return pf, nil
		}
	}
}
