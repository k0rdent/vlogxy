package handler

import (
	"github.com/k0rdent/vlogxy/internal/aggregation"
	"github.com/k0rdent/vlogxy/internal/parser"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

// field holds the accumulated values for a single stats field across backends,
// along with the aggregation function name (empty for group-by fields).
type field struct {
	values []string
	fnName string
}

// mergeStatsPipe groups rows by the pipe's ByFields and aggregates each Func
// field using its corresponding aggregation strategy.
func mergeStatsPipe(pipe *parser.Pipe, rows []map[string]string) ([]map[string]string, error) {
	groups := make(map[common.JsonKey]map[string]*field)

	for _, record := range rows {
		keyMap := make(map[string]string)
		rowFields := make(map[string]*field)

		for _, name := range pipe.ByFields {
			v, ok := record[name]
			if !ok {
				continue
			}
			keyMap[name] = v
			rowFields[name] = &field{values: []string{v}}
		}

		for _, fn := range pipe.Funcs {
			v, ok := record[fn.ResultName]
			if !ok {
				continue
			}
			rowFields[fn.ResultName] = &field{fnName: fn.Name, values: []string{v}}
		}

		key, err := common.MakeJsonKey(keyMap)
		if err != nil {
			log.Errorf("flat stats: failed to make group key: %v", err)
			continue
		}

		group, exists := groups[key]
		if !exists {
			groups[key] = rowFields
			continue
		}

		for name, src := range rowFields {
			if dst := group[name]; dst == nil {
				group[name] = src
			} else {
				dst.values = append(dst.values, src.values...)
			}
		}
	}

	result := make([]map[string]string, 0, len(groups))
	for _, group := range groups {
		out := make(map[string]string, len(group))
		for name, f := range group {
			out[name] = aggregation.Aggregate(f.fnName, f.values)
		}
		result = append(result, out)
	}
	return result, nil
}
