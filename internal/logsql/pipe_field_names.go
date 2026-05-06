package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeFieldNames processes '| field_names' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#field_names-pipe
type pipeFieldNames struct {
	// resultName is an optional name of the column to write results to.
	// By default results are written into 'name' column.
	resultName string

	// if the filter is non-empty then only the field names containing the given filter substring are returned.
	filter string

	// if isFirstPipe is set, then there is no need in loading columnsHeader in writeBlock().
	isFirstPipe bool
}

func (pf *pipeFieldNames) String() string {
	s := "field_names"
	if pf.filter != "" {
		s += " filter " + quoteTokenIfNeeded(pf.filter)
	}
	if pf.resultName != "name" {
		s += " as " + quoteTokenIfNeeded(pf.resultName)
	}
	return s
}

func (pf *pipeFieldNames) Name() string {
	return "field_names"
}

func (pf *pipeFieldNames) updateNeededFields(f *prefixfilter.Filter) {
	if pf.isFirstPipe {
		f.Reset()
	} else {
		f.AddAllowFilter("*")
	}
}

func (pf *pipeFieldNames) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeFieldNames(lex *lexer) (pipe, error) {
	if !lex.isKeyword("field_names") {
		return nil, fmt.Errorf("expecting 'field_names'; got %q", lex.token)
	}
	lex.nextToken()

	filter := ""
	if lex.isKeyword("filter") {
		lex.nextToken()
		f, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot parse filter inside 'field_names' pipe: %w", err)
		}
		filter = f
	}

	resultName := "name"
	if lex.isKeyword("as") {
		lex.nextToken()
		name, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result name for 'field_names': %w", err)
		}
		resultName = name
	} else if !lex.isKeyword("", "|") {
		name, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result name for 'field_names': %w", err)
		}
		resultName = name
	}

	pf := &pipeFieldNames{
		resultName: resultName,
		filter:     filter,
	}
	return pf, nil
}
