package merger

import (
	"sort"

	"github.com/k0rdent/vlogxy/internal/parser"
)

// PipeMerger applies merge logic for a single pipe type.
// It receives the pipe configuration and the accumulated rows from all backends,
// and returns the merged rows.
type PipeMerger func(pipe *parser.Pipe, rows []map[string]string) ([]map[string]string, error)

// registeredMerger pairs a PipeMerger with a processing order.
// Mergers with a lower Order value run first.
// When Exclusive is true, the presence of this merger in a query causes all
// non-exclusive mergers to be skipped — use this for pipes that completely
// replace the output format (e.g. facets).
// When KeepWithExclusive is true, the merger is kept even when an exclusive
// merger is present — use this for post-processing pipes like sort.
type registeredMerger struct {
	Order             int
	Exclusive         bool
	KeepWithExclusive bool
	Merge             PipeMerger
}

// pipeMergers maps pipe names to their merge handlers.
// To add support for a new pipe type, register a registeredMerger here.
// Pipes without a registered merger are passed through unchanged.
var pipeMergers = map[string]registeredMerger{
	"stats":  {Order: 1, Merge: mergeStatsPipe},
	"facets": {Order: 2, Exclusive: true, Merge: mergeFacetPipe},
	"sort":   {Order: 3, KeepWithExclusive: true, Merge: mergeSortPipe},
}

// pipeTask holds a resolved pipe together with its merger and sort key.
type pipeTask struct {
	Pipe              *parser.Pipe
	Merge             PipeMerger
	exclusive         bool
	keepWithExclusive bool
	order             int
}

// OrderedPipeTasks returns the subset of pipes that have registered mergers,
// sorted ascending by Order. Pipes sharing the same Order retain their
// original query position (stable sort).
// If any exclusive merger is present, all non-exclusive mergers are dropped.
func OrderedPipeTasks(pipes []*parser.Pipe) []pipeTask {
	var tasks []pipeTask
	hasExclusive := false
	for _, p := range pipes {
		if rm, ok := pipeMergers[p.Name]; ok {
			tasks = append(tasks, pipeTask{Pipe: p, order: rm.Order, exclusive: rm.Exclusive, keepWithExclusive: rm.KeepWithExclusive, Merge: rm.Merge})
			if rm.Exclusive {
				hasExclusive = true
			}
		}
	}

	if hasExclusive {
		filtered := tasks[:0]
		for _, t := range tasks {
			if t.exclusive || t.keepWithExclusive {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].order < tasks[j].order
	})
	return tasks
}
