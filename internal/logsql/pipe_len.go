package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeLen processes '| len ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#len-pipe
type pipeLen struct {
	fieldName   string
	resultField string
}

func (pl *pipeLen) String() string {
	s := "len(" + quoteTokenIfNeeded(pl.fieldName) + ")"
	if !isMsgFieldName(pl.resultField) {
		s += " as " + quoteTokenIfNeeded(pl.resultField)
	}
	return s
}

func (pl *pipeLen) Name() string {
	return "len"
}

func (pl *pipeLen) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchString(pl.resultField) {
		pf.AddDenyFilter(pl.resultField)
		pf.AddAllowFilter(pl.fieldName)
	}
}

func (pl *pipeLen) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeLen(lex *lexer) (pipe, error) {
	if !lex.isKeyword("len") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "len")
	}
	lex.nextToken()

	fieldName, err := parseFieldNameWithOptionalParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse field name for 'len' pipe: %w", err)
	}

	// parse optional 'as ...` part
	resultField := "_msg"
	if lex.isKeyword("as") {
		lex.nextToken()
	}
	if !lex.isKeyword("|", ")", "") {
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result field after 'len(%s)': %w", quoteTokenIfNeeded(fieldName), err)
		}
		resultField = field
	}

	pl := &pipeLen{
		fieldName:   fieldName,
		resultField: resultField,
	}

	return pl, nil
}

func parseFieldNameWithOptionalParens(lex *lexer) (string, error) {
	hasParens := false
	if lex.isKeyword("(") {
		lex.nextToken()
		hasParens = true
	}
	fieldName, err := parseFieldName(lex)
	if err != nil {
		return "", err
	}
	if hasParens {
		if !lex.isKeyword(")") {
			return "", fmt.Errorf("missing ')' after '%s'", quoteTokenIfNeeded(fieldName))
		}
		lex.nextToken()
	}
	return fieldName, nil
}
