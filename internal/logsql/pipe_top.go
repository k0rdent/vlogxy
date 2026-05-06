package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeTopDefaultLimit is the default number of entries pipeTop returns.
const pipeTopDefaultLimit = 10

// pipeTop processes '| top ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#top-pipe
type pipeTop struct {
	// fields contains field names for returning top values for.
	byFields []string

	// limit is the number of top (byFields) sets to return.
	limit uint64

	// limitStr is string representation of the limit.
	limitStr string

	// the number of hits per each unique value is returned in this field.
	hitsFieldName string

	// if rankFieldName isn't empty, then the rank per each unique value is returned in this field.
	rankFieldName string
}

func (pt *pipeTop) String() string {
	s := "top"
	if pt.limit != pipeTopDefaultLimit {
		s += " " + pt.limitStr
	}
	s += " by (" + fieldNamesString(pt.byFields) + ")"
	if pt.hitsFieldName != "hits" {
		s += " hits as " + quoteTokenIfNeeded(pt.hitsFieldName)
	}
	if pt.rankFieldName != "" {
		s += rankFieldNameString(pt.rankFieldName)
	}
	return s
}

func (pt *pipeTop) Name() string {
	return "top"
}

func (pt *pipeTop) updateNeededFields(pf *prefixfilter.Filter) {
	pf.Reset()
	pf.AddAllowFilters(pt.byFields)
}

func (pt *pipeTop) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeTop(lex *lexer) (pipe, error) {
	if !lex.isKeyword("top") {
		return nil, fmt.Errorf("expecting 'top'; got %q", lex.token)
	}
	lex.nextToken()

	limit := uint64(pipeTopDefaultLimit)
	limitStr := ""
	if isNumberPrefix(lex.token) {
		limitF, s, err := parseNumber(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse N in 'top': %w", err)
		}
		if limitF < 1 {
			return nil, fmt.Errorf("value N in 'top %s' must be integer bigger than 0", s)
		}
		limit = uint64(limitF)
		limitStr = s
	}

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
	} else if !lex.isKeyword("hits", "rank", ")", "|", "") {
		bfs, err := parseCommaSeparatedFields(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'by ...': %w", err)
		}
		byFields = bfs
	} else if needFields {
		return nil, fmt.Errorf("missing fields after 'by'")
	}
	if len(byFields) == 0 {
		return nil, fmt.Errorf("expecting at least a single field in 'by(...)'")
	}

	pt := &pipeTop{
		byFields:      byFields,
		limit:         limit,
		limitStr:      limitStr,
		hitsFieldName: "hits",
	}

	for {
		switch {
		case lex.isKeyword("hits"):
			lex.nextToken()
			if lex.isKeyword("as") {
				lex.nextToken()
			}
			s, err := lex.nextCompoundToken()
			if err != nil {
				return nil, fmt.Errorf("cannot parse 'hits' name: %w", err)
			}
			pt.hitsFieldName = s
		case lex.isKeyword("rank"):
			rankFieldName, err := parseRankFieldName(lex)
			if err != nil {
				return nil, fmt.Errorf("cannot parse rank field name in [%s]: %w", pt, err)
			}
			pt.rankFieldName = getUniqueResultName(rankFieldName, byFields)
		default:
			pt.hitsFieldName = getUniqueResultName(pt.hitsFieldName, byFields)
			return pt, nil
		}
	}
}

func parseRankFieldName(lex *lexer) (string, error) {
	if !lex.isKeyword("rank") {
		return "", fmt.Errorf("unexpected token: %q; want 'rank'", lex.token)
	}
	lex.nextToken()

	rankFieldName := "rank"
	if lex.isKeyword("as") {
		lex.nextToken()
		if lex.isKeyword("", "|", ")", "(") {
			return "", fmt.Errorf("missing rank name")
		}
	}
	if !lex.isKeyword("", "|", ")", "limit") {
		s, err := parseFieldName(lex)
		if err != nil {
			return "", err
		}
		rankFieldName = s
	}
	return rankFieldName, nil
}

func rankFieldNameString(rankFieldName string) string {
	s := " rank"
	if rankFieldName != "rank" {
		s += " as " + quoteTokenIfNeeded(rankFieldName)
	}
	return s
}
