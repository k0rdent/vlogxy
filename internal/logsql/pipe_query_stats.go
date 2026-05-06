package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeQueryStats implements '| query_stats' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#query_stats-pipe
type pipeQueryStats struct {
}

func (ps *pipeQueryStats) String() string {
	return "query_stats"
}

func (ps *pipeQueryStats) Name() string {
	return ps.String()
}

func (ps *pipeQueryStats) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("*")
}

func (ps *pipeQueryStats) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeQueryStats(lex *lexer) (pipe, error) {
	if !lex.isKeyword("query_stats") {
		return nil, fmt.Errorf("expecting 'query_stats'; got %q", lex.token)
	}
	lex.nextToken()

	ps := &pipeQueryStats{}

	return ps, nil
}
