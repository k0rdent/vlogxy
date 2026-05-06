package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeSample implements '| sample ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#limit-sample
type pipeSample struct {
	// sample shows how many rows on average must be skipped during sampling
	sample uint64
}

func (ps *pipeSample) String() string {
	return fmt.Sprintf("sample %d", ps.sample)
}

func (ps *pipeSample) Name() string {
	return "sample"
}

func (ps *pipeSample) updateNeededFields(_ *prefixfilter.Filter) {
	// nothing to do
}

func (ps *pipeSample) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeSample(lex *lexer) (pipe, error) {
	if !lex.isKeyword("sample") {
		return nil, fmt.Errorf("expecting 'sample'; got %q", lex.token)
	}
	lex.nextToken()

	sampleStr, err := lex.nextCompoundToken()
	if err != nil {
		return nil, fmt.Errorf("cannot read 'sample': %w", err)
	}

	sample, err := parseUint(sampleStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse sample from %q: %w", sampleStr, err)
	}

	if sample <= 0 {
		return nil, fmt.Errorf("unexpected sample=%d; it must be bigger than 0", sample)
	}

	ps := &pipeSample{
		sample: sample,
	}
	return ps, nil
}
