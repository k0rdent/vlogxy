package logstorage

import (
	"fmt"
	"strings"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeSort processes '| sort ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#sort-pipe
type pipeSort struct {
	// byFields contains field names for sorting from 'by(...)' clause.
	byFields []*bySortField

	// whether to apply descending order
	isDesc bool

	// how many results to skip
	offset uint64

	// how many results to return
	//
	// if zero, then all the results are returned
	limit uint64

	// The name of the field to store the row rank.
	rankFieldName string

	// partitionByFields contains fields for partitioning the sorted rows.
	partitionByFields []string
}

func (ps *pipeSort) String() string {
	s := "sort"
	if len(ps.byFields) > 0 {
		a := make([]string, len(ps.byFields))
		for i, bf := range ps.byFields {
			a[i] = bf.String()
		}
		s += " by (" + strings.Join(a, ", ") + ")"
	}
	if ps.isDesc {
		s += " desc"
	}

	if len(ps.partitionByFields) > 0 {
		s += " partition by (" + fieldNamesString(ps.partitionByFields) + ")"
	}

	if ps.offset > 0 {
		s += fmt.Sprintf(" offset %d", ps.offset)
	}
	if ps.limit > 0 {
		s += fmt.Sprintf(" limit %d", ps.limit)
	}

	if ps.rankFieldName != "" {
		s += rankFieldNameString(ps.rankFieldName)
	}

	return s
}

func (ps *pipeSort) Name() string {
	return "sort"
}

func (ps *pipeSort) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchNothing() {
		// There is no need in fetching any fields, since all of them are ignored by the caller.
		return
	}

	if ps.rankFieldName != "" {
		pf.AddDenyFilter(ps.rankFieldName)
	}

	if len(ps.byFields) == 0 {
		pf.AddAllowFilter("*")
	} else {
		for _, bf := range ps.byFields {
			pf.AddAllowFilter(bf.name)
		}
	}

	pf.AddAllowFilters(ps.partitionByFields)
}

func (ps *pipeSort) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

// SortField is an exported description of a single sort key.
type SortField struct {
	// Name is the field name to sort by.
	Name string
	// IsDesc indicates descending order for this field.
	IsDesc bool
}

// SortFields returns the ordered list of sort keys for this pipe.
// When the list is empty the whole row set is sorted as a single group
// using the pipe-level IsDesc flag.
func (ps *pipeSort) SortFields() []SortField {
	fields := make([]SortField, len(ps.byFields))
	for i, bf := range ps.byFields {
		fields[i] = SortField{Name: bf.name, IsDesc: bf.isDesc}
	}
	return fields
}

// SortIsDesc returns true when the pipe-level descending flag is set
// (applies when no per-field directions are specified).
func (ps *pipeSort) SortIsDesc() bool { return ps.isDesc }

// SortLimit returns the row limit (0 means unlimited).
func (ps *pipeSort) SortLimit() uint64 { return ps.limit }

// PipeSortAccessor is implemented by sort pipes and lets the parser layer
// read sort configuration without importing the concrete unexported type.
type PipeSortAccessor interface {
	SortFields() []SortField
	SortIsDesc() bool
	SortLimit() uint64
}

func parsePipeSort(lex *lexer) (pipe, error) {
	if !lex.isKeyword("sort") && !lex.isKeyword("order") {
		return nil, fmt.Errorf("expecting 'sort' or 'order'; got %q", lex.token)
	}
	lex.nextToken()

	var ps pipeSort
	if lex.isKeyword("by", "(") {
		if lex.isKeyword("by") {
			lex.nextToken()
		}
		bfs, err := parseBySortFields(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by' clause: %w", err)
		}
		ps.byFields = bfs
	}

	switch {
	case lex.isKeyword("desc"):
		lex.nextToken()
		ps.isDesc = true
	case lex.isKeyword("asc"):
		lex.nextToken()
	}

	for {
		switch {
		case lex.isKeyword("offset"):
			n, err := parseOffset(lex)
			if err != nil {
				return nil, err
			}
			if ps.offset > 0 {
				return nil, fmt.Errorf("duplicate 'offset'; the previous one is %d; the new one is %d", ps.offset, n)
			}
			ps.offset = n
		case lex.isKeyword("limit"):
			n, err := parseLimit(lex)
			if err != nil {
				return nil, err
			}
			if ps.limit > 0 {
				return nil, fmt.Errorf("duplicate 'limit'; the previous one is %d; the new one is %d", ps.limit, n)
			}
			ps.limit = n
		case lex.isKeyword("rank"):
			rankFieldName, err := parseRankFieldName(lex)
			if err != nil {
				return nil, fmt.Errorf("cannot read rank field name: %s", err)
			}
			ps.rankFieldName = rankFieldName
		case lex.isKeyword("partition"):
			if len(ps.partitionByFields) > 0 {
				return nil, fmt.Errorf("duplicate 'partition by'")
			}
			lex.nextToken()
			if lex.isKeyword("by") {
				lex.nextToken()
			}
			fields, err := parseFieldNamesInParens(lex)
			if err != nil {
				return nil, fmt.Errorf("cannot parse 'partition by' args: %w", err)
			}
			ps.partitionByFields = fields
		default:
			if len(ps.partitionByFields) > 0 && ps.limit <= 0 {
				return nil, fmt.Errorf("missing 'limit' for 'partition by'")
			}
			return &ps, nil
		}
	}
}

// bySortField represents 'by (...)' part of the pipeSort.
type bySortField struct {
	// the name of the field to sort
	name string

	// whether the sorting for the given field in descending order
	isDesc bool
}

func (bf *bySortField) String() string {
	s := quoteTokenIfNeeded(bf.name)
	if bf.isDesc {
		s += " desc"
	}
	return s
}

func parseBySortFields(lex *lexer) ([]*bySortField, error) {
	if !lex.isKeyword("(") {
		return nil, fmt.Errorf("missing `(`")
	}
	var bfs []*bySortField
	for {
		lex.nextToken()
		if lex.isKeyword(")") {
			lex.nextToken()
			return bfs, nil
		}
		fieldName, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse field name: %w", err)
		}
		bf := &bySortField{
			name: fieldName,
		}
		switch {
		case lex.isKeyword("desc"):
			lex.nextToken()
			bf.isDesc = true
		case lex.isKeyword("asc"):
			lex.nextToken()
		}
		bfs = append(bfs, bf)
		switch {
		case lex.isKeyword(")"):
			lex.nextToken()
			return bfs, nil
		case lex.isKeyword(","):
		default:
			return nil, fmt.Errorf("unexpected token: %q; expecting ',' or ')'", lex.token)
		}
	}
}

func parseLimit(lex *lexer) (uint64, error) {
	if !lex.isKeyword("limit") {
		return 0, fmt.Errorf("expecting 'limit'; got %q", lex.token)
	}
	lex.nextToken()

	limitStr, err := lex.nextCompoundToken()
	if err != nil {
		return 0, fmt.Errorf("cannot parse 'limit': %s", err)
	}

	n, ok := tryParseUint64(limitStr)
	if !ok {
		return 0, fmt.Errorf("cannot parse %q as number in the 'limit'", limitStr)
	}

	return n, nil
}

func parseOffset(lex *lexer) (uint64, error) {
	if !lex.isKeyword("offset") {
		return 0, fmt.Errorf("expecting 'offset'; got %q", lex.token)
	}
	lex.nextToken()

	limitStr, err := lex.nextCompoundToken()
	if err != nil {
		return 0, fmt.Errorf("cannot parse 'offset': %s", err)
	}

	n, ok := tryParseUint64(limitStr)
	if !ok {
		return 0, fmt.Errorf("cannot parse %q as number in the 'offset'", limitStr)
	}

	return n, nil
}
