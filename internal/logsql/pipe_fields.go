package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFields implements '| fields ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#fields-pipe
type pipeFields struct {
	// fieldFilters contains list of filters for fields to fetch
	fieldFilters []string
}

func (pf *pipeFields) String() string {
	if len(pf.fieldFilters) == 0 {
		logger.Panicf("BUG: pipeFields must contain at least a single field filter")
	}
	return "fields " + fieldNamesString(pf.fieldFilters)
}

func (pf *pipeFields) Name() string {
	return "fields"
}

func (pf *pipeFields) updateNeededFields(f *prefixfilter.Filter) {
	fOrig := f.Clone()
	f.Reset()

	for _, filter := range pf.fieldFilters {
		if fOrig.MatchStringOrWildcard(filter) {
			f.AddAllowFilter(filter)
		}
	}
}

func (pf *pipeFields) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeFields(lex *lexer) (pipe, error) {
	if !lex.isKeyword("fields", "keep") {
		return nil, fmt.Errorf("expecting 'fields'; got %q", lex.token)
	}
	lex.nextToken()

	fieldFilters, err := parseCommaSeparatedFields(lex)
	if err != nil {
		return nil, err
	}
	pf := &pipeFields{
		fieldFilters: fieldFilters,
	}
	return pf, nil
}

func parseCommaSeparatedFields(lex *lexer) ([]string, error) {
	var fields []string
	for {
		field, err := parseFieldFilter(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse field name: %w", err)
		}
		fields = append(fields, field)
		if !lex.isKeyword(",") {
			return fields, nil
		}
		lex.nextToken()
	}
}
