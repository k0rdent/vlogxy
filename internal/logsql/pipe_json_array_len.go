package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeJSONArrayLen processes '| json_array_len ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#json_array_len-pipe
type pipeJSONArrayLen struct {
	fieldName   string
	resultField string
}

func (pl *pipeJSONArrayLen) String() string {
	s := "json_array_len(" + quoteTokenIfNeeded(pl.fieldName) + ")"
	if !isMsgFieldName(pl.resultField) {
		s += " as " + quoteTokenIfNeeded(pl.resultField)
	}
	return s
}

func (pl *pipeJSONArrayLen) Name() string {
	return "json_array_len"
}

func (pl *pipeJSONArrayLen) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchString(pl.resultField) {
		pf.AddDenyFilter(pl.resultField)
		pf.AddAllowFilter(pl.fieldName)
	}
}

func (pl *pipeJSONArrayLen) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeJSONArrayLen(lex *lexer) (pipe, error) {
	if !lex.isKeyword("json_array_len") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "json_array_len")
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

	pl := &pipeJSONArrayLen{
		fieldName:   fieldName,
		resultField: resultField,
	}

	return pl, nil
}
