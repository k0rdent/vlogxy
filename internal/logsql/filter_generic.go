package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterGeneric applies the given filter f to the given fieldName
type filterGeneric struct {
	// fieldName is the name of the field to apply f to.
	//
	// It may end with '*' if isWildcard is true.
	fieldName string

	// isWildcard indicates whether fieldName is a wildcard ending with '*'
	//
	// In this case f is applied to all the fields with the given fieldName prefix until the first match.
	isWildcard bool

	// f is the filter to apply.
	f fieldFilter
}

func newFilterGeneric(fieldName string, f fieldFilter) *filterGeneric {
	if prefixfilter.IsWildcardFilter(fieldName) {
		return &filterGeneric{
			fieldName:  fieldName,
			isWildcard: true,
			f:          f,
		}
	}

	fieldNameCanonical := getCanonicalColumnName(fieldName)
	return &filterGeneric{
		fieldName: fieldNameCanonical,
		f:         f,
	}
}

func (fg *filterGeneric) visitSubqueries(visitFunc func(q *Query)) {
	switch t := fg.f.(type) {
	case *filterContainsAll:
		t.values.q.visitSubqueries(visitFunc)
	case *filterContainsAny:
		t.values.q.visitSubqueries(visitFunc)
	case *filterIn:
		t.values.q.visitSubqueries(visitFunc)
	default:
		// nothing to do
	}
}

// String returns string representation of the fg.
func (fg *filterGeneric) String() string {
	if !fg.isWildcard {
		return quoteFieldNameIfNeeded(fg.fieldName) + fg.f.String()
	}

	return quoteFieldFilterIfNeeded(fg.fieldName) + ":" + fg.f.String()
}

func (fg *filterGeneric) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(fg.fieldName)
}

func quoteFieldNameIfNeeded(s string) string {
	if isMsgFieldName(s) {
		return ""
	}
	return quoteTokenIfNeeded(s) + ":"
}

func isMsgFieldName(fieldName string) bool {
	return fieldName == "" || fieldName == "_msg"
}
