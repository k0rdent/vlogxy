package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterEqField matches if the given fields have equivalent values.
//
// Example LogsQL: `fieldName:eq_field(otherField)`
type filterEqField struct {
	fieldName      string
	otherFieldName string
}

func newFilterEqField(fieldName, otherFieldName string) *filterEqField {
	return &filterEqField{
		fieldName:      getCanonicalColumnName(fieldName),
		otherFieldName: getCanonicalColumnName(otherFieldName),
	}
}

func (fe *filterEqField) String() string {
	return fmt.Sprintf("%seq_field(%s)", quoteFieldNameIfNeeded(fe.fieldName), quoteTokenIfNeeded(fe.otherFieldName))
}

func (fe *filterEqField) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(fe.fieldName)
	pf.AddAllowFilter(fe.otherFieldName)
}
