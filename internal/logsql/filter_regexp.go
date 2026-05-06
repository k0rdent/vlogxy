package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/regexutil"
)

// filterRegexp matches the given regexp
//
// Example LogsQL: `re("regexp")`
type filterRegexp struct {
	re *regexutil.Regex
}

func newFilterRegexp(fieldName string, re *regexutil.Regex) *filterGeneric {
	fp := &filterRegexp{
		re: re,
	}
	return newFilterGeneric(fieldName, fp)
}

func (fr *filterRegexp) String() string {
	return fmt.Sprintf("~%s", quoteTokenIfNeeded(fr.re.String()))
}
