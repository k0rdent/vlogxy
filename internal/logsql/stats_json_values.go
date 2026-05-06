package logstorage

import (
	"fmt"
	"strings"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type statsJSONValues struct {
	// fieldFilters contains field filters for fields to select from logs.
	fieldFilters []string

	// sortFields contains optional fields for sorting the selected logs.
	//
	// if sortFields is empty, then the selected logs aren't sorted.
	sortFields []*bySortField

	// limit contains an optional limit on the number of logs to select.
	//
	// if limit==0, then all the logs are selected.
	limit uint64
}

func (sv *statsJSONValues) Name() string {
	return "json_values"
}

func (sv *statsJSONValues) String() string {
	s := sv.Name() + "(" + fieldNamesString(sv.fieldFilters) + ")"

	if len(sv.sortFields) > 0 {
		a := make([]string, len(sv.sortFields))
		for i, sf := range sv.sortFields {
			a[i] = sf.String()
		}
		s += fmt.Sprintf(" sort by (%s)", strings.Join(a, ", "))
	}

	if sv.limit > 0 {
		s += fmt.Sprintf(" limit %d", sv.limit)
	}
	return s
}

func (sv *statsJSONValues) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilters(sv.fieldFilters)

	for _, sf := range sv.sortFields {
		pf.AddAllowFilter(sf.name)
	}
}

func parseStatsJSONValues(lex *lexer) (statsFunc, error) {
	fieldFilters, err := parseStatsFuncFieldFilters(lex, "json_values")
	if err != nil {
		return nil, err
	}

	sv := &statsJSONValues{
		fieldFilters: fieldFilters,
	}

	if lex.isKeyword("sort", "order") {
		lex.nextToken()
		if lex.isKeyword("by") {
			lex.nextToken()
		}
		sfs, err := parseBySortFields(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'sort': %w", err)
		}
		sv.sortFields = sfs
	}

	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		sv.limit = n
	}
	return sv, nil
}
