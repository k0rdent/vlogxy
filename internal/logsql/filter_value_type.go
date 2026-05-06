package logstorage

import (
	"fmt"
)

// filterValueType filters field entries by value type.
//
// For example, the following filter returns all the logs with uint64 fieldName:
//
//	fieldName:value_type("uint64")
type filterValueType struct {
	valueType string
}

func newFilterValueType(fieldName, valueType string) *filterGeneric {
	fv := &filterValueType{
		valueType: valueType,
	}
	return newFilterGeneric(fieldName, fv)
}

func (fv *filterValueType) String() string {
	return fmt.Sprintf("value_type(%s)", quoteTokenIfNeeded(fv.valueType))
}
