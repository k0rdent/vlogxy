package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsAvg struct {
	fieldFilters []string
}

func (sa *statsAvg) Name() string {
	return "avg"
}

func (sa *statsAvg) String() string {
	return sa.Name() + "(" + fieldNamesString(sa.fieldFilters) + ")"
}

func (sa *statsAvg) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sa.fieldFilters)
}

func parseStatsAvg(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "avg")
	if err != nil {
		return nil, err
	}
	sa := &statsAvg{
		fieldFilters: fieldFilters,
	}
	return sa, nil
}

func parseStatsFuncFields(lex *lexer, funcName string) ([]string, error) {
	if !lex.isKeyword(funcName) {
		return nil, fmt.Errorf("unexpected func; got %q; want %q", lex.token, funcName)
	}
	lex.nextToken()
	fields, err := parseFieldFiltersInParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q args: %w", funcName, err)
	}

	// Check that all the selected fields are real fields
	for _, f := range fields {
		if prefixfilter.IsWildcardFilter(f) {
			return nil, fmt.Errorf("unexpected wildcard filter %q inside %s()", f, funcName)
		}
	}

	return fields, nil
}

func parseStatsFuncArgs(lex *lexer, funcName string) ([]string, error) {
	if !lex.isKeyword(funcName) {
		return nil, fmt.Errorf("unexpected func; got %q; want %q", lex.token, funcName)
	}
	lex.nextToken()
	fields, err := parseFieldNamesInParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q args: %w", funcName, err)
	}
	return fields, nil
}

func parseStatsFuncFieldFilters(lex *lexer, funcName string) ([]string, error) {
	if !lex.isKeyword(funcName) {
		return nil, fmt.Errorf("unexpected func; got %q; want %q", lex.token, funcName)
	}
	lex.nextToken()
	fields, err := parseFieldFiltersInParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q args: %w", funcName, err)
	}
	if len(fields) == 0 {
		fields = []string{"*"}
	}
	return fields, nil
}
