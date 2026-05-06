package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsHistogram struct {
	fieldName string
}

func (sh *statsHistogram) Name() string {
	return "histogram"
}

func (sh *statsHistogram) String() string {
	return sh.Name() + "(" + quoteTokenIfNeeded(sh.fieldName) + ")"
}

func (sh *statsHistogram) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(sh.fieldName)
}

func parseStatsHistogram(lex *lexer) (statsFunc, error) {
	fields, err := parseStatsFuncFields(lex, "histogram")
	if err != nil {
		return nil, fmt.Errorf("cannot parse field name: %w", err)
	}
	if len(fields) != 1 {
		return nil, fmt.Errorf("'histogram' accepts only a single field; got %d fields", len(fields))
	}

	sh := &statsHistogram{
		fieldName: fields[0],
	}
	return sh, nil
}
