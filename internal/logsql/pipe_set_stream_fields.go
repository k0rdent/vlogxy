package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeSetStreamFields processes '| set_stream_fields ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#set_stream_fields-pipe
type pipeSetStreamFields struct {
	streamFieldFilters []string

	// iff is an optional filter for skipping setting stream fields
	iff *ifFilter
}

func (ps *pipeSetStreamFields) String() string {
	s := "set_stream_fields"
	if ps.iff != nil {
		s += " " + ps.iff.String()
	}
	s += " " + fieldNamesString(ps.streamFieldFilters)
	return s
}

func (ps *pipeSetStreamFields) Name() string {
	return "set_stream_fields"
}

func (ps *pipeSetStreamFields) updateNeededFields(f *prefixfilter.Filter) {
	if !f.MatchString("_stream") {
		return
	}

	if ps.iff != nil {
		f.AddAllowFilters(ps.iff.allowFilters)
	} else {
		f.AddDenyFilter("_stream")
	}
	f.AddAllowFilters(ps.streamFieldFilters)
}

func (ps *pipeSetStreamFields) visitSubqueries(visitFunc func(q *Query)) {
	ps.iff.visitSubqueries(visitFunc)
}

func parsePipeSetStreamFields(lex *lexer) (pipe, error) {
	if !lex.isKeyword("set_stream_fields") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "set_stream_fields")
	}
	lex.nextToken()

	// parse optional if (...)
	var iff *ifFilter
	if lex.isKeyword("if") {
		f, err := parseIfFilter(lex)
		if err != nil {
			return nil, err
		}
		iff = f
	}

	// Parse stream fields
	streamFieldFilters, err := parseCommaSeparatedFields(lex)
	if err != nil {
		return nil, err
	}

	ps := &pipeSetStreamFields{
		streamFieldFilters: streamFieldFilters,

		iff: iff,
	}

	return ps, nil
}
