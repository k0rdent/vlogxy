package logstorage

import (
	"fmt"
)

// filterExactPrefix matches the exact prefix.
//
// Example LogsQL: `="foo bar"*`
type filterExactPrefix struct {
	prefix string
}

func newFilterExactPrefix(fieldName, prefix string) *filterGeneric {
	fe := &filterExactPrefix{
		prefix: prefix,
	}
	return newFilterGeneric(fieldName, fe)
}

func (fep *filterExactPrefix) String() string {
	return fmt.Sprintf("=%s*", quoteTokenIfNeeded(fep.prefix))
}
