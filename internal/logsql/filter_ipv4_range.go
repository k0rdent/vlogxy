package logstorage

import (
	"fmt"
)

// filterIPv4Range matches the given ipv4 range [minValue..maxValue].
//
// Example LogsQL: `ipv4_range(127.0.0.1, 127.0.0.255)`
type filterIPv4Range struct {
	minValue uint32
	maxValue uint32
}

func newFilterIPv4Range(fieldName string, minValue, maxValue uint32) *filterGeneric {
	fr := &filterIPv4Range{
		minValue: minValue,
		maxValue: maxValue,
	}
	return newFilterGeneric(fieldName, fr)
}

func (fr *filterIPv4Range) String() string {
	minValue := marshalIPv4String(nil, fr.minValue)
	maxValue := marshalIPv4String(nil, fr.maxValue)
	return fmt.Sprintf("ipv4_range(%s, %s)", minValue, maxValue)
}
