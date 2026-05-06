package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUnion processes '| union ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#union-pipe
type pipeUnion struct {
	// q is a query for obtaining results to add after all the input results are processed.
	//
	// q is nil if rows is non-nil.
	q *Query

	// rows contains rows to add after processing all the input results.
	//
	// rows are obtained either by executing q at initUnionQuery
	// or they can be put inline in the union pipe via the following syntax:
	//
	//     union rows({row1}, ... {rowN})
	//
	rows [][]Field
}

func (pu *pipeUnion) String() string {
	var dst []byte
	dst = append(dst, "union "...)

	if pu.rows != nil {
		dst = marshalRows(dst, pu.rows)
	} else {
		dst = append(dst, '(')
		dst = append(dst, pu.q.String()...)
		dst = append(dst, ')')
	}

	return string(dst)
}

func (pu *pipeUnion) Name() string {
	return "union"
}

func (pu *pipeUnion) visitSubqueries(visitFunc func(q *Query)) {
	pu.q.visitSubqueries(visitFunc)
}

func (pu *pipeUnion) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func parsePipeUnion(lex *lexer) (pipe, error) {
	if !lex.isKeyword("union") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "union")
	}
	lex.nextToken()

	var q *Query
	var rows [][]Field
	var err error
	if lex.isKeyword("rows") {
		rows, err = parseRows(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse rows inside 'union': %w", err)
		}
	} else {
		q, err = parseQueryInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse subquery inside 'union': %w", err)
		}
	}

	pu := &pipeUnion{
		q:    q,
		rows: rows,
	}
	return pu, nil
}
