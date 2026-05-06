package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeTimeAdd processes '| time_add ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#time_add-pipe
type pipeTimeAdd struct {
	field string

	offset    int64
	offsetStr string
}

func (pa *pipeTimeAdd) String() string {
	s := "time_add " + pa.offsetStr
	if pa.field != "_time" {
		s += " at " + quoteTokenIfNeeded(pa.field)
	}
	return s
}

func (pa *pipeTimeAdd) Name() string {
	return "time_add"
}

func (pa *pipeTimeAdd) updateNeededFields(_ *prefixfilter.Filter) {
	// do nothing
}

func (pa *pipeTimeAdd) visitSubqueries(_ func(q *Query)) {
	// do nothing
}

func parsePipeTimeAdd(lex *lexer) (pipe, error) {
	if !lex.isKeyword("time_add") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "time_add")
	}
	lex.nextToken()

	offset, offsetStr, err := parseDuration(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse offset: %w", err)
	}

	// Parse optional field
	field := "_time"
	if lex.isKeyword("at") {
		lex.nextToken()
		fieldName, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot read field name: %w", err)
		}
		field = fieldName
	}

	pa := &pipeTimeAdd{
		field:     field,
		offset:    -offset,
		offsetStr: offsetStr,
	}

	return pa, nil
}
