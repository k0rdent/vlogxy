package logstorage

import (
	"fmt"
)

// filterIn matches any exact value from the values map.
//
// Example LogsQL: `in("foo", "bar baz")`
type filterIn struct {
	values inValues
}

func (fi *filterIn) String() string {
	args := fi.values.String()
	return fmt.Sprintf("in(%s)", args)
}
