package logstorage

var maxStringRangeValue = string([]byte{255, 255, 255, 255})

// filterStringRange matches tie given string range [minValue..maxValue)
//
// Note that the minValue is included in the range, while the maxValue isn't included in the range.
// This simplifies querying distinct log sets with string_range(A, B), string_range(B, C), etc.
//
// Example LogsQL: `string_range(minValue, maxValue)`
type filterStringRange struct {
	minValue string
	maxValue string

	stringRepr string
}

func newFilterStringRange(fieldName, minValue, maxValue, stringRepr string) *filterGeneric {
	fr := &filterStringRange{
		minValue: minValue,
		maxValue: maxValue,

		stringRepr: stringRepr,
	}
	return newFilterGeneric(fieldName, fr)
}

func (fr *filterStringRange) String() string {
	return fr.stringRepr
}
