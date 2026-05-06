package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/internal/parser"
	"github.com/k0rdent/vlogxy/pkg/common"
)

// StatsResponse represents the response structure for stats queries
type StatsResponse struct {
	Data struct {
		ResultType string        `json:"resultType"`
		Result     []StatsSeries `json:"result"`
	} `json:"data"`
	Status string `json:"status"`
}

// StatsSeries represents a single time series in stats response
type StatsSeries struct {
	Metric map[string]string `json:"metric"`
	Value  common.ValuePair  `json:"value"`
}

// Stats handles aggregation of stats query responses from multiple backends.
// When pipes are provided it converts responses to flat rows, runs orderedPipeTasks,
// then reconstructs a StatsResponse.
type Stats struct {
	pipes []*parser.Pipe
}

// NewStatsQuery creates a Stats aggregator with pipe-based merging.
func NewStatsQuery(pipes []*parser.Pipe) interfaces.ResponseAggregator[StatsResponse] {
	return &Stats{pipes: pipes}
}

func (s *Stats) ParseResponse(resp *http.Response) (StatsResponse, error) {
	return common.DecodeJSONResponse[StatsResponse](resp)
}

func (s *Stats) Merge(responses []StatsResponse) ([]byte, error) {
	return s.mergeWithPipes(responses)
}

// mergeWithPipes converts each StatsSeries to a flat row, runs orderedPipeTasks,
// then reconstructs a StatsResponse. __ts__ is injected into stats-pipe ByFields
// (same as StatsRange) so mergeStatsPipe groups correctly.
func (s *Stats) mergeWithPipes(responses []StatsResponse) ([]byte, error) {
	// Collect the query timestamp from the first available series.
	var queryTimestamp float64
	for _, resp := range responses {
		for _, series := range resp.Data.Result {
			if ts, ok := series.Value[0].(float64); ok {
				queryTimestamp = ts
				break
			}
		}
		if queryTimestamp > 0 {
			break
		}
	}

	// Collect result names from the stats pipe.
	resultNames := make(map[string]struct{})
	for _, p := range s.pipes {
		if p.Name == "stats" {
			for _, fn := range p.Funcs {
				resultNames[fn.ResultName] = struct{}{}
			}
			break
		}
	}

	// Convert all backend series to flat rows.
	var rows []map[string]string
	for _, resp := range responses {
		for _, series := range resp.Data.Result {
			row := make(map[string]string, len(series.Metric)+2)
			for k, v := range series.Metric {
				row[k] = v
			}
			row[rangeTimestampField] = strconv.FormatFloat(queryTimestamp, 'f', -1, 64)
			resultName := rangeSeriesResultName(series.Metric, resultNames)
			if resultName != "" {
				if val, ok := series.Value[1].(string); ok {
					row[resultName] = val
				}
			}
			rows = append(rows, row)
		}
	}

	// Inject __ts__ into each stats pipe's ByFields so mergeStatsPipe groups correctly.
	modifiedPipes := make([]*parser.Pipe, len(s.pipes))
	for i, p := range s.pipes {
		if p.Name == "stats" {
			cp := *p
			cp.ByFields = append(append([]string{}, p.ByFields...), rangeTimestampField)
			modifiedPipes[i] = &cp
		} else {
			modifiedPipes[i] = p
		}
	}

	var err error
	for _, task := range orderedPipeTasks(modifiedPipes) {
		rows, err = task.merge(task.pipe, rows)
		if err != nil {
			return nil, err
		}
	}

	var result StatsResponse
	if len(responses) > 0 {
		result.Status = responses[0].Status
		result.Data.ResultType = responses[0].Data.ResultType
	}

	result.Data.Result = make([]StatsSeries, 0, len(rows))
	for _, row := range rows {
		metric := make(map[string]string, len(row))
		valStr := ""
		for k, v := range row {
			if k == rangeTimestampField {
				continue
			} else if _, isResult := resultNames[k]; isResult {
				valStr = v
			} else {
				metric[k] = v
			}
		}
		result.Data.Result = append(result.Data.Result, StatsSeries{
			Metric: metric,
			Value:  common.ValuePair{queryTimestamp, valStr},
		})
	}

	return json.Marshal(result)
}
