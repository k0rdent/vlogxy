package logstorage

import (
	"fmt"
)

// filterPrefix matches the given prefix.
//
// Example LogsQL: `prefix*` or `"some prefix"*`
//
// A special case `*` matches non-empty value for the given `fieldName` field
type filterPrefix struct {
	prefix string
}

func newFilterPrefix(fieldName, prefix string) *filterGeneric {
	fp := &filterPrefix{
		prefix: prefix,
	}
	return newFilterGeneric(fieldName, fp)
}

func (fp *filterPrefix) String() string {
	if fp.prefix == "" {
		return "*"
	}
	return fmt.Sprintf("%s*", quoteTokenIfNeeded(fp.prefix))
}
