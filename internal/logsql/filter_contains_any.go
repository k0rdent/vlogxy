package logstorage

import (
	"fmt"
)

// filterContainsAny matches any value from the values.
//
// Example LogsQL: `contains_any("foo", "bar baz")`
type filterContainsAny struct {
	values inValues
}

func (fi *filterContainsAny) String() string {
	args := fi.values.String()
	return fmt.Sprintf("contains_any(%s)", args)
}
