package logstorage

import (
	"fmt"
)

// filterContainsAll matches logs containing all the given values.
//
// Example LogsQL: `contains_all("foo", "bar baz")`
type filterContainsAll struct {
	values inValues
}

func (fi *filterContainsAll) String() string {
	args := fi.values.String()
	return fmt.Sprintf("contains_all(%s)", args)
}
