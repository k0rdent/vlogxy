package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeBlockStats processes '| block_stats ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#block_stats-pipe
type pipeBlockStats struct {
}

func (ps *pipeBlockStats) String() string {
	return "block_stats"
}

func (ps *pipeBlockStats) Name() string {
	return ps.String()
}

func (ps *pipeBlockStats) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func (ps *pipeBlockStats) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("*")
}

func parsePipeBlockStats(lex *lexer) (pipe, error) {
	if !lex.isKeyword("block_stats") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "block_stats")
	}
	lex.nextToken()

	ps := &pipeBlockStats{}

	return ps, nil
}
