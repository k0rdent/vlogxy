package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeUnpackWords processes '| unpack_words ...' pipe.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#unpack_words-pipe
type pipeUnpackWords struct {
	// field to unpack words from
	srcField string

	// field to put the unpacked words
	dstField string

	// whether to drop duplicate words
	dropDuplicates bool
}

func (pu *pipeUnpackWords) String() string {
	s := "unpack_words"
	if pu.srcField != "_msg" {
		s += " from " + quoteTokenIfNeeded(pu.srcField)
	}
	if pu.dstField != pu.srcField {
		s += " as " + quoteTokenIfNeeded(pu.dstField)
	}
	if pu.dropDuplicates {
		s += " drop_duplicates"
	}
	return s
}

func (pu *pipeUnpackWords) Name() string {
	return "unpack_words"
}

func (pu *pipeUnpackWords) visitSubqueries(_ func(q *Query)) {
	// do nothing
}

func (pu *pipeUnpackWords) updateNeededFields(pf *prefixfilter.Filter) {
	if pf.MatchString(pu.dstField) {
		pf.AddDenyFilter(pu.dstField)
		pf.AddAllowFilter(pu.srcField)
	}
}

func parsePipeUnpackWords(lex *lexer) (pipe, error) {
	if !lex.isKeyword("unpack_words") {
		return nil, fmt.Errorf("unexpected token: %q; want %q", lex.token, "unpack_words")
	}
	lex.nextToken()

	srcField := "_msg"
	if !lex.isKeyword("drop_duplicates", "as", ")", "|", "") {
		if lex.isKeyword("from") {
			lex.nextToken()
		}
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse srcField name: %w", err)
		}
		srcField = field
	}

	dstField := srcField
	if !lex.isKeyword("drop_duplicates", ")", "|", "") {
		if lex.isKeyword("as") {
			lex.nextToken()
		}
		field, err := parseFieldName(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse dstField name: %w", err)
		}
		dstField = field
	}

	dropDuplicates := false
	if lex.isKeyword("drop_duplicates") {
		lex.nextToken()
		dropDuplicates = true
	}

	pu := &pipeUnpackWords{
		srcField: srcField,
		dstField: dstField,

		dropDuplicates: dropDuplicates,
	}

	return pu, nil
}
