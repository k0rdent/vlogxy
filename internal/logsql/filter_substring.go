package logstorage

import (
	"fmt"
)

// filterSubstring filters field entries by substring match.
//
// An empty substring matches any string.
type filterSubstring struct {
	substring string
}

func newFilterSubstring(fieldName, substring string) *filterGeneric {
	fs := &filterSubstring{
		substring: substring,
	}
	return newFilterGeneric(fieldName, fs)
}

func (fs *filterSubstring) String() string {
	return fmt.Sprintf("*%s*", quoteTokenIfNeeded(fs.substring))
}
