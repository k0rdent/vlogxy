package logstorage

import (
	"fmt"
	"slices"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeJoin processes '| join ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#join-pipe
type pipeJoin struct {
	// byFields contains fields to use for join on q results
	byFields []string

	// q is a query for obtaining results for joining
	//
	// it is set to nil if rows is non-nil.
	q *Query

	// rows contains inline log rows for joining
	//
	// rows are obtained either by executing q at initJoinMap
	// or they can be put inline in the join pipe via the following syntax:
	//
	//     join by (...) rows({row1}, ... {rowN})
	//
	rows [][]Field

	// The join is performed as INNER JOIN if isInner is set.
	// Otherwise the join is performed as LEFT JOIN.
	isInner bool

	// prefix is the prefix to add to log fields from q query
	prefix string
}

func (pj *pipeJoin) String() string {
	var dst []byte
	dst = append(dst, "join by ("...)
	dst = append(dst, fieldNamesString(pj.byFields)...)
	dst = append(dst, ") "...)

	if pj.rows != nil {
		dst = marshalRows(dst, pj.rows)
	} else {
		dst = append(dst, '(')
		dst = append(dst, pj.q.String()...)
		dst = append(dst, ')')
	}

	if pj.isInner {
		dst = append(dst, " inner"...)
	}
	if pj.prefix != "" {
		dst = append(dst, " prefix "...)
		dst = append(dst, quoteTokenIfNeeded(pj.prefix)...)
	}
	return string(dst)
}

func (pj *pipeJoin) Name() string {
	return "join"
}

func (pj *pipeJoin) visitSubqueries(visitFunc func(q *Query)) {
	if pj.q != nil {
		pj.q.visitSubqueries(visitFunc)
	}
}

func (pj *pipeJoin) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(pj.byFields)
}

func parsePipeJoin(lex *lexer) (pipe, error) {
	if !lex.isKeyword("join") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "join")
	}
	lex.nextToken()

	// parse by (...)
	if lex.isKeyword("by", "on") {
		lex.nextToken()
	}

	byFields, err := parseFieldNamesInParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'by(...)' at 'join': %w", err)
	}
	if len(byFields) == 0 {
		return nil, fmt.Errorf("'by(...)' at 'join' must contain at least a single field")
	}
	if slices.Contains(byFields, "*") {
		return nil, fmt.Errorf("join by '*' isn't supported")
	}

	var q *Query
	var rows [][]Field
	if lex.isKeyword("rows") {
		rows, err = parseRows(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse rows inside 'join': %w", err)
		}
	} else {
		q, err = parseQueryInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse subquery inside 'join': %w", err)
		}
	}

	pj := &pipeJoin{
		byFields: byFields,
		q:        q,
		rows:     rows,
	}

	if lex.isKeyword("inner") {
		lex.nextToken()
		pj.isInner = true
	}

	if lex.isKeyword("prefix") {
		lex.nextToken()
		prefix, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot read prefix for [%s]: %w", pj, err)
		}
		pj.prefix = prefix

		if !pj.isInner && lex.isKeyword("inner") {
			lex.nextToken()
			pj.isInner = true
		}
	}

	return pj, nil
}

func marshalRows(dst []byte, rows [][]Field) []byte {
	if len(rows) == 0 {
		return append(dst, "rows()"...)
	}

	dst = append(dst, "rows("...)
	for _, row := range rows {
		dst = MarshalFieldsToJSON(dst, row)
		dst = append(dst, ',')
	}
	dst[len(dst)-1] = ')'

	return dst
}

func parseRows(lex *lexer) ([][]Field, error) {
	if !lex.isKeyword("rows") {
		return nil, fmt.Errorf("missing 'rows' prefix")
	}
	lex.nextToken()

	if !lex.isKeyword("(") {
		return nil, fmt.Errorf("missing '(' after 'rows' prefix")
	}
	lex.nextToken()

	// It is important to do not return nil rows here, since the caller depends on non-nil rows.
	rows := [][]Field{}

	for !lex.isKeyword(")") {
		row, err := parseRow(lex)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)

		if lex.isKeyword(",") {
			lex.nextToken()
		}
	}
	lex.nextToken()

	return rows, nil
}

func parseRow(lex *lexer) ([]Field, error) {
	if !lex.isKeyword("{") {
		return nil, fmt.Errorf("missing '{'; got %q instead", lex.token)
	}
	lex.nextToken()

	var fields []Field

	for !lex.isKeyword("}") {
		name := lex.token
		lex.nextToken()

		if !lex.isKeyword(":", "=") {
			return nil, fmt.Errorf("missing ':' or '=' after %q; got [%s] instead", name, lex.token)
		}
		lex.nextToken()

		value, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot read value after %q: %w", name, err)
		}

		fields = append(fields, Field{
			Name:  name,
			Value: value,
		})

		if lex.isKeyword("}") {
			break
		}

		if lex.isKeyword(",") {
			lex.nextToken()
		}
	}
	lex.nextToken()

	return fields, nil
}
