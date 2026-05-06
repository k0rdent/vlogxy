package logstorage

import (
	"strings"
)

// inValues keeps values for in(...), contains_any(...) and contains_all(...) filters
type inValues struct {
	values []string

	// If q is non-nil, then values must be populated from q before filter execution.
	q *Query

	// qFieldName must be set to field name for obtaining values from if q is non-nil.
	qFieldName string
}

func (iv *inValues) String() string {
	if iv.q != nil {
		return iv.q.String()
	}
	values := iv.values
	a := make([]string, len(values))
	for i, value := range values {
		a[i] = quoteTokenIfNeeded(value)
	}
	return strings.Join(a, ",")
}
