package handler

import (
	"sort"
	"strconv"

	logstorage "github.com/k0rdent/vlogxy/internal/logsql"
	"github.com/k0rdent/vlogxy/internal/parser"
)

// mergeSortPipe re-sorts the merged rows according to the sort pipe's field
// list and applies the row limit when set. It works after any other merger
// (stats, facets, …) because it only cares about the final row values.
func mergeSortPipe(pipe *parser.Pipe, rows []map[string]string) ([]map[string]string, error) {
	fields := pipe.SortFields
	pipeDesc := pipe.SortIsDesc

	sort.SliceStable(rows, func(i, j int) bool {
		return sortLess(rows[i], rows[j], fields, pipeDesc)
	})

	if pipe.SortLimit > 0 && uint64(len(rows)) > pipe.SortLimit {
		rows = rows[:pipe.SortLimit]
	}

	return rows, nil
}

// sortLess returns true when row a should come before row b.
func sortLess(a, b map[string]string, fields []logstorage.SortField, pipeDesc bool) bool {
	if len(fields) == 0 {
		// No per-field spec: compare all shared string keys lexicographically.
		// The pipe-level isDesc flag inverts the whole comparison.
		for k, va := range a {
			vb, ok := b[k]
			if !ok {
				continue
			}
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
		if f.IsDesc {
			return !less
		}
		return less
	}
	return false
}

// compareValues compares two string values. When both parse as numbers the
// numeric order is used; otherwise plain string comparison is applied.
func compareValues(a, b string) bool {
	na, errA := strconv.ParseFloat(a, 64)
	nb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		return na < nb
	}
	return a < b
}
