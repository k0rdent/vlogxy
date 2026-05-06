package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsQuantile struct {
	fieldFilters []string

	phi    float64
	phiStr string
}

func (sq *statsQuantile) Name() string {
	return "quantile"
}

func (sq *statsQuantile) String() string {
	s := sq.Name() + "(" + sq.phiStr
	if !prefixfilter.MatchAll(sq.fieldFilters) {
		s += ", " + fieldNamesString(sq.fieldFilters)
	}
	s += ")"
	return s
}

func (sq *statsQuantile) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sq.fieldFilters)
}

func parseStatsQuantile(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "quantile")
	if err != nil {
		return nil, err
	}
	if len(fieldFilters) == 0 {
		return nil, fmt.Errorf("missing phi arg at 'quantile'")
	}

	// Parse phi
	phiStr := fieldFilters[0]
	phi, ok := tryParseFloat64(phiStr)
	if !ok {
		return nil, fmt.Errorf("phi arg in 'quantile' must be floating point number; got %q", phiStr)
	}
	if phi < 0 || phi > 1 {
		return nil, fmt.Errorf("phi arg in 'quantile' must be in the range [0..1]; got %q", phiStr)
	}

	// Parse fields
	fieldFilters = fieldFilters[1:]
	if len(fieldFilters) == 0 {
		fieldFilters = []string{"*"}
	}

	sq := &statsQuantile{
		fieldFilters: fieldFilters,

		phi:    phi,
		phiStr: phiStr,
	}
	return sq, nil
}
