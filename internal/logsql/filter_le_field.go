package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterLeField matches if the fieldName field is smaller or equal to the otherFieldName field
//
// Example LogsQL: `fieldName:le_field(otherField)`
type filterLeField struct {
	fieldName      string
	otherFieldName string

	excludeEqualValues bool
}

func newFilterLeField(fieldName, otherFieldName string, excludeEqualValues bool) *filterLeField {
	return &filterLeField{
		fieldName:      getCanonicalColumnName(fieldName),
		otherFieldName: getCanonicalColumnName(otherFieldName),

		excludeEqualValues: excludeEqualValues,
	}
}

func (fe *filterLeField) String() string {
	funcName := "le_field"
	if fe.excludeEqualValues {
		funcName = "lt_field"
	}
	return fmt.Sprintf("%s%s(%s)", quoteFieldNameIfNeeded(fe.fieldName), funcName, quoteTokenIfNeeded(fe.otherFieldName))
}

func (fe *filterLeField) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(fe.fieldName)
	pf.AddAllowFilter(fe.otherFieldName)
}
