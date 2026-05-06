package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeHash processes '| hash ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#hash-pipe
type pipeHash struct {
	fieldName   string
	resultField string
}

func (ph *pipeHash) String() string {
	s := "hash(" + quoteTokenIfNeeded(ph.fieldName) + ")"
	if !isMsgFieldName(ph.resultField) {
		s += " as " + quoteTokenIfNeeded(ph.resultField)
	}
	return s
}

func (ph *pipeHash) Name() string {
	return "hash"
}

func (ph *pipeHash) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchString(ph.resultField) {
		pf.AddDenyFilter(ph.resultField)
		pf.AddAllowFilter(ph.fieldName)
	}
}

func (ph *pipeHash) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeHash(lex *lexer) (pipe, error) {
	if !lex.isKeyword("hash") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "hash")
	}
	lex.nextToken()

	fieldName, err := parseFieldNameWithOptionalParens(lex)
	if err != nil {
		return nil, fmt.Errorf("cannot parse field name for 'hash' pipe: %w", err)
	}

	// parse optional 'as ...` part
	resultField := "_msg"
	if lex.isKeyword("as") {
		lex.nextToken()
	}
	if !lex.isKeyword("|", ")", "") {
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse result field after 'hash(%s)': %w", quoteTokenIfNeeded(fieldName), err)
		}
		resultField = field
	}

	ph := &pipeHash{
		fieldName:   fieldName,
		resultField: resultField,
	}

	return ph, nil
}
