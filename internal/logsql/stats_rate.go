package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsRate struct {
	// stepSeconds must be updated by the caller before calling newStatsProcessor().
	stepSeconds float64
}

func (sr *statsRate) Name() string {
	return "rate"
}

func (sr *statsRate) String() string {
	return sr.Name() + "()"
}

func (sr *statsRate) updateNeededFields(_ *prefixfilter.Filter) {
	// There is no need in fetching any columns for rate() - the number of matching rows can be calculated as blockResult.rowsLen
}

func parseStatsRate(lex *lexer) (statsFunc, error) {
	fields, err := parseStatsFuncFields(lex, "rate")
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		return nil, fmt.Errorf("unexpected non-empty args for 'rate()' function: %q", fields)
	}
	sr := &statsRate{}
	return sr, nil
}
