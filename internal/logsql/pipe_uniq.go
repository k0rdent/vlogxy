package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUniq processes '| uniq ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#uniq-pipe
type pipeUniq struct {
	// fields contains field names for returning unique values
	byFields []string

	// if the filter is non-empty then only the values containing the given filter substring are returned.
	filter string

	// if hitsFieldName isn't empty, then the number of hits per each unique value is stored in this field.
	hitsFieldName string

	// limit is the maximum number of unique values to return.
	// If hitsFieldName != "" and the limit is exceeded, then all the hits are set to 0.
	limit uint64
}

func (pu *pipeUniq) String() string {
	s := "uniq by (" + fieldNamesString(pu.byFields) + ")"
	if pu.filter != "" {
		s += " filter " + quoteTokenIfNeeded(pu.filter)
	}
	if pu.hitsFieldName != "" {
		s += " with hits"
	}
	if pu.limit > 0 {
		s += fmt.Sprintf(" limit %d", pu.limit)
	}
	return s
}

func (pu *pipeUniq) Name() string {
	return "uniq"
}

func (pu *pipeUniq) updateNeededFields(pf *prefixfilter.Filter) {
	pf.Reset()
	pf.AddAllowFilters(pu.byFields)
}

func (pu *pipeUniq) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeUniq(lex *lexer) (pipe, error) {
	if !lex.isKeyword("uniq") {
		return nil, fmt.Errorf("expecting 'uniq'; got %q", lex.token)
	}
	lex.nextToken()

	needFields := false
	if lex.isKeyword("by") {
		lex.nextToken()
		needFields = true
	}

	var byFields []string
	if lex.isKeyword("(") {
		bfs, err := parseFieldNamesInParens(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by(...)': %w", err)
		}
		byFields = bfs
	} else if !lex.isKeyword("filter", "with", "hits", "limit", ")", "|", "") {
		bfs, err := parseCommaSeparatedFields(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by ...': %w", err)
		}
		byFields = bfs
	} else if needFields {
		return nil, fmt.Errorf("missing fields after 'by'")
	}
	if len(byFields) == 0 {
		return nil, fmt.Errorf("missing fields inside 'by(...)'")
	}

	pu := &pipeUniq{
		byFields: byFields,
	}

	if lex.isKeyword("filter") {
		lex.nextToken()
		f, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot parse filter inside 'uniq' pipe: %w", err)
		}
		pu.filter = f
		if len(byFields) != 1 && pu.filter != "" {
			return nil, fmt.Errorf("the 'filter %s' inside 'uniq' pipe cannot be applied to multiple fields %s", quoteTokenIfNeeded(pu.filter), byFields)
		}
	}

	if lex.isKeyword("with") {
		lex.nextToken()
		if !lex.isKeyword("hits") {
			return nil, fmt.Errorf("missing 'hits' after 'with'")
		}
	}
	if lex.isKeyword("hits") {
		lex.nextToken()
		pu.hitsFieldName = getUniqueResultName("hits", pu.byFields)
	}

	if lex.isKeyword("limit") {
		n, err := parseLimit(lex)
		if err != nil {
			return nil, err
		}
		pu.limit = n
	}

	return pu, nil
}
