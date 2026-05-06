package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeOffset implements '| offset ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#offset-pipe
type pipeOffset struct {
	offset uint64
}

func (po *pipeOffset) String() string {
	return fmt.Sprintf("offset %d", po.offset)
}

func (po *pipeOffset) Name() string {
	return "offset"
}

func (po *pipeOffset) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func (po *pipeOffset) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeOffset(lex *lexer) (pipe, error) {
	if !lex.isKeyword("offset", "skip") {
		return nil, fmt.Errorf("expecting 'offset' or 'skip'; got %q", lex.token)
	}
	op := lex.token
	lex.nextToken()

	token, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot parse '%s': %w", op, err)
	}
	n, err := parseUint(token)
	if err != nil {
		return nil, fmt.Errorf("cannot parse '%s %s': %w", op, token, err)
	}
	po := &pipeOffset{
		offset: n,
	}
	return po, nil
}
