package logstorage

// filterRange matches the given range [minValue..maxValue].
//
// Example LogsQL: `range(minValue, maxValue]`
type filterRange struct {
	minValue float64
	maxValue float64

	stringRepr string
}

func newFilterRange(fieldName string, minValue, maxValue float64, stringRepr string) *filterGeneric {
	fr := &filterRange{
		minValue: minValue,
		maxValue: maxValue,

		stringRepr: stringRepr,
	}
	return newFilterGeneric(fieldName, fr)
}

func (fr *filterRange) String() string {
	return fr.stringRepr
}
