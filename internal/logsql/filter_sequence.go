package logstorage

import (
	"fmt"
	"strings"
)

// filterSequence matches an ordered sequence of phrases
//
// Example LogsQL: `seq(foo, "bar baz")`
type filterSequence struct {
	phrases []string
}

func newFilterSequence(fieldName string, phrases []string) *filterGeneric {
	fs := &filterSequence{
		phrases: phrases,
	}
	return newFilterGeneric(fieldName, fs)
}

func (fs *filterSequence) String() string {
	phrases := fs.phrases
	a := make([]string, len(phrases))
	for i, phrase := range phrases {
		a[i] = quoteTokenIfNeeded(phrase)
	}
	return fmt.Sprintf("seq(%s)", strings.Join(a, ","))
}
