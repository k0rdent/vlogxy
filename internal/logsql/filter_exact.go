package logstorage

import (
	"fmt"
)

// filterExact matches the exact value.
//
// Example LogsQL: `exact("foo bar")` of `="foo bar"
type filterExact struct {
	value string
}

func newFilterExact(fieldName, value string) *filterGeneric {
	fe := &filterExact{
		value: value,
	}
	return newFilterGeneric(fieldName, fe)
}

func (fe *filterExact) String() string {
	return fmt.Sprintf("=%s", quoteTokenIfNeeded(fe.value))
}
