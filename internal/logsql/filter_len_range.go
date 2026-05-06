package logstorage

// filterLenRange matches field values with the length in the given range [minLen, maxLen].
//
// Example LogsQL: `len_range(10, 20)`
type filterLenRange struct {
	minLen uint64
	maxLen uint64

	stringRepr string
}

func newFilterLenRange(fieldName string, minLen, maxLen uint64, stringRepr string) *filterGeneric {
	fr := &filterLenRange{
		minLen: minLen,
		maxLen: maxLen,

		stringRepr: stringRepr,
	}
	return newFilterGeneric(fieldName, fr)
}

func (fr *filterLenRange) String() string {
	return "len_range" + fr.stringRepr
}
