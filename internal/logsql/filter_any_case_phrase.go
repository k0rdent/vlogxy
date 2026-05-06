package logstorage

import (
	"fmt"
)

// filterAnyCasePhrase filters field entries by case-insensitive phrase match.
//
// An example LogsQL query: `i(word)` or `i("word1 ... wordN")`
type filterAnyCasePhrase struct {
	phrase string
}

func newFilterAnyCasePhrase(fieldName, phrase string) *filterGeneric {
	fp := &filterAnyCasePhrase{
		phrase: phrase,
	}
	return newFilterGeneric(fieldName, fp)
}

func (fp *filterAnyCasePhrase) String() string {
	return fmt.Sprintf("i(%s)", quoteTokenIfNeeded(fp.phrase))
}
