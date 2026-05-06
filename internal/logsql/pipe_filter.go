package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFilter processes '| filter ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#filter-pipe
type pipeFilter struct {
	// f is a filter to apply to the written rows.
	f filter
}

func (pf *pipeFilter) String() string {
	return "filter " + pf.f.String()
}

func (pf *pipeFilter) Name() string {
	return "filter"
}

func (pf *pipeFilter) updateNeededFields(f *prefixfilter.Filter) {
	pf.f.updateNeededFields(f)
}

func (pf *pipeFilter) visitSubqueries(visitFunc func(q *Query)) {
	visitSubqueriesInFilter(pf.f, visitFunc)
}

func parsePipeFilter(lex *lexer) (pipe, error) {
	return parsePipeFilterExt(lex, true)
}

func parsePipeFilterNoFilterKeyword(lex *lexer) (pipe, error) {
	return parsePipeFilterExt(lex, false)
}

func parsePipeFilterExt(lex *lexer, needFilterKeyword bool) (pipe, error) {
	if needFilterKeyword {
		if !lex.isKeyword("filter", "where") {
			return nil, fmt.Errorf("expecting 'filter' or 'where'; got %q", lex.token)
		}
		lex.nextToken()
	}

	f, err := parseFilter(lex, needFilterKeyword)
	if err != nil {
		return nil, fmt.Errorf("cannot parse 'filter': %w", err)
	}

	pf := &pipeFilter{
		f: f,
	}
	return pf, nil
}
