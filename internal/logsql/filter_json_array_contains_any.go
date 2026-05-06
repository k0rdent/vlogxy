package logstorage

import (
	"fmt"
	"strings"
)

// filterJSONArrayContainsAny matches if the JSON array in the given field contains the given value.
//
// Example LogsQL: `tags:json_array_contains_any("prod","dev")`
type filterJSONArrayContainsAny struct {
	values []string
}

func newFilterJSONArrayContainsAny(fieldName string, values []string) *filterGeneric {
	fa := &filterJSONArrayContainsAny{
		values: values,
	}
	return newFilterGeneric(fieldName, fa)
}

func (fa *filterJSONArrayContainsAny) String() string {
	a := make([]string, len(fa.values))
	for i, v := range fa.values {
		a[i] = quoteTokenIfNeeded(v)
	}
	args := strings.Join(a, ",")
	return fmt.Sprintf("json_array_contains_any(%s)", args)
}
