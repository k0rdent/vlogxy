package logstorage

// filterPhrase filters field entries by phrase match (aka full text search).
//
// A phrase consists of any number of words with delimiters between them.
//
// An empty phrase matches only an empty string.
// A single-word phrase is the simplest LogsQL query: `word`
//
// Multi-word phrase is expressed as `"word1 ... wordN"` in LogsQL.
//
// A special case `""` matches any log entry without the given `fieldName` field.
type filterPhrase struct {
	phrase string
}

func newFilterPhrase(fieldName, phrase string) *filterGeneric {
	fp := &filterPhrase{
		phrase: phrase,
	}
	return newFilterGeneric(fieldName, fp)
}

func (fp *filterPhrase) String() string {
	return quoteTokenIfNeeded(fp.phrase)
}
