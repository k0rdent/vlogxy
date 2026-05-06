package logstorage

import (
	"fmt"
	"strings"
)

// filterEqualsCommonCase matches words and phrases where every capital letter
// can be replaced with a small letter, plus all capital words.
//
// Example LogsQL: `equals_common_case("Error")` is equivalent to in("Error", "error", "ERROR")
type filterEqualsCommonCase struct {
	phrases []string

	equalsAny filterIn
}

func newFilterEqualsCommonCase(fieldName string, phrases []string) (*filterGeneric, error) {
	commonCasePhrases, err := getCommonCasePhrases(phrases)
	if err != nil {
		return nil, err
	}

	fi := &filterEqualsCommonCase{
		phrases: phrases,
	}
	fi.equalsAny.values.values = commonCasePhrases

	fg := newFilterGeneric(fieldName, fi)
	return fg, nil
}

func (fi *filterEqualsCommonCase) String() string {
	a := make([]string, len(fi.phrases))
	for i, phrase := range fi.phrases {
		a[i] = quoteTokenIfNeeded(phrase)
	}
	phrases := strings.Join(a, ",")
	return fmt.Sprintf("equals_common_case(%s)", phrases)
}
