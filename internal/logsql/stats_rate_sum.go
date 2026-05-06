package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsRateSum struct {
	ss *statsSum

	// stepSeconds must be updated by the caller before calling newStatsProcessor().
	stepSeconds float64
}

func (sr *statsRateSum) Name() string {
	return "rate_sum"
}

func (sr *statsRateSum) String() string {
	return sr.Name() + "(" + fieldNamesString(sr.ss.fieldFilters) + ")"
}

func (sr *statsRateSum) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sr.ss.fieldFilters)
}

func parseStatsRateSum(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "rate_sum")
	if err != nil {
		return nil, err
	}
	sr := &statsRateSum{
		ss: &statsSum{
			fieldFilters: fieldFilters,
		},
	}
	return sr, nil
}
