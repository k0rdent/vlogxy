package parser

import (
	"fmt"

	"github.com/k0rdent/vlogxy/internal/logsql"
)

type Pipe struct {
	// Name is the name of the pipe, e.g. "stats", "unpack_words", etc.
	Name string
	// Raw is the raw query string for the pipe.
	Raw string
	// ByFields is the list of fields used in the BY clause of a stats pipe, e.g. "service", "env", etc.
	ByFields []string
	// Funcs is the list of functions used in the pipe, e.g. "count()", "sum(value)", etc.
	Funcs []*Func
	// SortFields is the ordered list of sort keys for a sort pipe.
	// Each entry carries the field name and whether it is sorted descending.
	SortFields []logsql.SortField
	// SortIsDesc is the pipe-level descending flag (applies when SortFields is empty).
	SortIsDesc bool
	// SortLimit is the row limit declared in the sort pipe (0 = unlimited).
	SortLimit uint64
}

type Func struct {
	// Name is the function name, e.g. "count", "sum", "quantile", etc.
	Name string
	// FullName is the full function name including arguments, e.g. "count()", "sum(value)", etc.
	FullName string
	// Raw is the raw query string for the function, e.g. "count() as total_count", "sum(k8s.container.restart_count) as restart_sum", etc.
	Raw string
	// ResultName is the name of the field where the function result will be stored, e.g. "total_logs", "total_restarts", etc.
	ResultName string
}

func ParseQuery(query string) ([]*Pipe, error) {
	q, err := logsql.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	pipes := make([]*Pipe, 0)
	for _, pipe := range q.GetPipes() {
		funcs := make([]*Func, 0)
		byFields := make([]string, 0)

		if stats, ok := pipe.(*logsql.PipeStats); ok {
			for _, f := range stats.ByFields() {
				byFields = append(byFields, f.Name())
			}

			for _, f := range stats.Funcs() {
				funcs = append(funcs, &Func{
					Name:       f.Func().Name(),
					FullName:   f.Func().String(),
					Raw:        f.String(),
					ResultName: f.ResultName(),
				})
			}
		}

		if totalStats, ok := pipe.(*logsql.PipeRunningStats); ok && totalStats.IsTotal() {
			byFields = append(byFields, totalStats.ByFields()...)

			for _, f := range totalStats.Funcs() {
				funcs = append(funcs, &Func{
					Name:       f.FuncName(),
					FullName:   f.FuncString(),
					Raw:        fmt.Sprintf("%s as %s", f.FuncString(), f.ResultName()),
					ResultName: f.ResultName(),
				})
			}
		}

		var sortFields []logsql.SortField
		var sortIsDesc bool
		var sortLimit uint64

		if s, ok := pipe.(*logsql.PipeSort); ok {
			byFields = s.ByFields()
			sortFields = s.SortFields()
			sortIsDesc = s.SortIsDesc()
			sortLimit = s.SortLimit()
			funcs = nil
		}

		pipes = append(pipes, &Pipe{
			Raw:        pipe.String(),
			Name:       pipe.Name(),
			ByFields:   byFields,
			Funcs:      funcs,
			SortFields: sortFields,
			SortIsDesc: sortIsDesc,
			SortLimit:  sortLimit,
		})
	}
	return pipes, nil
}
