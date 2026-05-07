package merger

import (
	"math"
	"sort"
	"strconv"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/stringsutil"
	"github.com/k0rdent/vlogxy/internal/logsql"
	"github.com/k0rdent/vlogxy/internal/parser"
)

// mergeSortPipe re-sorts the merged rows according to the sort pipe's field
// list and applies the row limit when set. It works after any other merger
// (stats, facets, …) because it only cares about the final row values.
func mergeSortPipe(pipe *parser.Pipe, rows []map[string]string) ([]map[string]string, error) {
	sort.SliceStable(rows, func(i, j int) bool {
		return sortLess(rows[i], rows[j], pipe.SortFields, pipe.SortIsDesc)
	})

	if pipe.SortLimit > 0 && uint64(len(rows)) > pipe.SortLimit {
		rows = rows[:pipe.SortLimit]
	}

	return rows, nil
}

// sortLess returns true when row a should come before row b.
func sortLess(a, b map[string]string, fields []logsql.SortField, pipeDesc bool) bool {
	if len(fields) == 0 {
		// No explicit sort keys: compare all fields in sorted key order so the
		// result is deterministic, then honour the pipe-level direction flag.
		for _, k := range unionKeys(a, b) {
			va, vb := a[k], b[k]
			if va == vb {
				continue
			}
			less := compareValues(va, vb)
			if pipeDesc {
				return !less
			}
			return less
		}
		return false
	}

	for _, f := range fields {
		va := a[f.Name]
		vb := b[f.Name]
		if va == vb {
			continue
		}
		less := compareValues(va, vb)
		if f.IsDesc || pipeDesc {
			return !less
		}
		return less
	}
	return false
}

// compareValues compares two string values using the same precedence as
// upstream VictoriaLogs: int64 → float64 → natural-sort string.
func compareValues(a, b string) bool {
	// Try integer comparison first.
	ia, errA := strconv.ParseInt(a, 10, 64)
	ib, errB := strconv.ParseInt(b, 10, 64)
	if errA == nil && errB == nil {
		return ia < ib
	}

	// Try float comparison next.
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil && !math.IsNaN(fa) && !math.IsNaN(fb) {
		return fa < fb
	}

	// Fall back to natural sort (matches VictoriaLogs sort-pipe behaviour).
	return stringsutil.LessNatural(a, b)
}

// unionKeys returns the sorted union of keys from both rows. Using a stable
// key ordering ensures that sortLess is a consistent comparator when no
// explicit sort fields are given.
func unionKeys(a, b map[string]string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
