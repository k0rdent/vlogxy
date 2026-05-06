package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeBlocksCount processes '| blocks_count' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#blocks_count-pipe
type pipeBlocksCount struct {
	// resultName is an optional name of the column to write results to.
	// By default results are written into 'blocks_count' column.
	resultName string
}

func (pc *pipeBlocksCount) String() string {
	s := "blocks_count"
	if pc.resultName != "blocks_count" {
		s += " as " + quoteTokenIfNeeded(pc.resultName)
	}
	return s
}

func (pc *pipeBlocksCount) Name() string {
	return "blocks_count"
}

func (pc *pipeBlocksCount) updateNeededFields(pf *prefixfilter.Filter) {
	pf.Reset()
}

func (pc *pipeBlocksCount) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeBlocksCount(lex *lexer) (pipe, error) {
	if !lex.isKeyword("blocks_count") {
		return nil, fmt.Errorf("expecting 'blocks_count'; got %q", lex.token)
	}
	lex.nextToken()

	resultName := "blocks_count"
	if lex.isKeyword("as") {
		lex.nextToken()
		name, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result name for 'blocks_count': %w", err)
		}
		resultName = name
	} else if !lex.isKeyword("", "|") {
		name, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result name for 'blocks_count': %w", err)
		}
		resultName = name
	}

	pc := &pipeBlocksCount{
		resultName: resultName,
	}
	return pc, nil
}
