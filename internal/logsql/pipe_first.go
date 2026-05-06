package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFirst processes '| first ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#first-pipe
type pipeFirst struct {
	ps *pipeSort
}

func (pf *pipeFirst) String() string {
	return pipeLastFirstString(pf.ps)
}

func (pf *pipeFirst) Name() string {
	return "first"
}

func (pf *pipeFirst) updateNeededFields(f *prefixfilter.Filter) {
	pf.ps.updateNeededFields(f)
}

func (pf *pipeFirst) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeFirst(lex *lexer) (pipe, error) {
	if !lex.isKeyword("first") {
		return nil, fmt.Errorf("expecting 'first'; got %q", lex.token)
	}
	lex.nextToken()

	ps, err := parsePipeLastFirst(lex)
	if err != nil {
		return nil, err
	}
	pf := &pipeFirst{
		ps: ps,
	}
	return pf, nil
}
