package logstorage

import (
	"fmt"
	"net/netip"
)

// filterIPv6Range matches the given ipv6 range [minValue..maxValue].
//
// Example LogsQL: `ipv6_range(::1, ::2)`
type filterIPv6Range struct {
	minValue [16]byte
	maxValue [16]byte
}

func newFilterIPv6Range(fieldName string, minValue, maxValue [16]byte) *filterGeneric {
	fr := &filterIPv6Range{
		minValue: minValue,
		maxValue: maxValue,
	}
	return newFilterGeneric(fieldName, fr)
}

func (fr *filterIPv6Range) String() string {
	minValue := netip.AddrFrom16(fr.minValue).String()
	maxValue := netip.AddrFrom16(fr.maxValue).String()
	return fmt.Sprintf("ipv6_range(%s, %s)", minValue, maxValue)
}
