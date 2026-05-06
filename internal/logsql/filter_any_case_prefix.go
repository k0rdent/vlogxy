package logstorage

import (
	"fmt"
)

// filterAnyCasePrefix matches the given prefix in lower, upper and mixed case.
//
// Example LogsQL: `i(prefix*)` or `i("some prefix"*)`
//
// A special case `i(*)` equals to `*` and matches non-empty value.
type filterAnyCasePrefix struct {
	prefix string
}

func newFilterAnyCasePrefix(fieldName, prefix string) *filterGeneric {
	fp := &filterAnyCasePrefix{
		prefix: prefix,
	}
	return newFilterGeneric(fieldName, fp)
}

func (fp *filterAnyCasePrefix) String() string {
	if fp.prefix == "" {
		return "i(*)"
	}
	return fmt.Sprintf("i(%s*)", quoteTokenIfNeeded(fp.prefix))
}
