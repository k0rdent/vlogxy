package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeGenerateSequence implements '| generate_sequence' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#generate_sequence-pipe
type pipeGenerateSequence struct {
	// n is the number of rows to generate in the sequence
	n uint64
}

func (pg *pipeGenerateSequence) String() string {
	return fmt.Sprintf("generate_sequence %d", pg.n)
}

func (pg *pipeGenerateSequence) Name() string {
	return "generate_sequence"
}

func (pg *pipeGenerateSequence) updateNeededFields(pf *prefixfilter.Filter) {
	pf.Reset()
}

func (pg *pipeGenerateSequence) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeGenerateSequence(lex *lexer) (pipe, error) {
	if !lex.isKeyword("generate_sequence") {
		return nil, fmt.Errorf("expecting 'generate_sequence'; got %q", lex.token)
	}
	lex.nextToken()

	if !isNumberPrefix(lex.token) {
		return nil, fmt.Errorf("expecting the number of items to generate in 'generate_sequence' pipe; got %q", lex.token)
	}
	nF, s, err := parseNumber(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse N in 'generate_sequence': %w", err)
	}
	if nF < 1 {
		return nil, fmt.Errorf("value N in 'generate_sequence %s' must be integer bigger than 0", s)
	}
	n := uint64(nF)

	pg := &pipeGenerateSequence{
		n: n,
	}
	return pg, nil
}
