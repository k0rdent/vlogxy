package handler

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/internal/parser"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

const rangeTimestampField = "__ts__"

// StatsRangeResponse represents the response structure for stats range queries
type StatsRangeResponse struct {
	Data struct {
		ResultType string             `json:"resultType"`
		Result     []StatsRangeSeries `json:"result"`
	} `json:"data"`
	Status string `json:"status"`
}

// StatsRangeSeries represents a time series with multiple values
type StatsRangeSeries struct {
	Metric map[string]string  `json:"metric"`
	Values []common.ValuePair `json:"values"`
}

// StatsRange handles aggregation of stats range query responses from multiple backends.
type StatsRange struct {
	pipes []*parser.Pipe
}

// NewStatsRangeQuery creates a StatsRange aggregator with pipe-based merging.
func NewStatsRangeQuery(pipes []*parser.Pipe) interfaces.ResponseAggregator[StatsRangeResponse] {
	return &StatsRange{pipes: pipes}
}

func (s *StatsRange) ParseResponse(resp *http.Response) (StatsRangeResponse, error) {
	return common.DecodeJSONResponse[StatsRangeResponse](resp)
}

func (s *StatsRange) Merge(responses []StatsRangeResponse) ([]byte, error) {
	return s.mergeWithPipes(responses)
}

// mergeWithPipes flattens each StatsRangeSeries into one row per (metric, timestamp):
//
//	{ label1: "v1", ..., "__ts__": "1234567890", resultName: "100" }
//
// It injects __ts__ into the stats pipe ByFields so that mergeStatsPipe groups by
// (by_fields + timestamp) and then reconstructs the StatsRangeResponse matrix.
func (s *StatsRange) mergeWithPipes(responses []StatsRangeResponse) ([]byte, error) {
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

	// Flatten: one row per (series, timestamp).
	var rows []map[string]string
	for _, resp := range responses {
		for _, series := range resp.Data.Result {
			for _, v := range series.Values {
				ts, ok := v[0].(float64)
				if !ok {
					log.Errorf("stats range: failed to parse timestamp")
					continue
				}
				val, ok := v[1].(string)
				if !ok {
					log.Errorf("stats range: failed to parse value")
					continue
				}

				row := make(map[string]string, len(series.Metric)+2)
				for k, mv := range series.Metric {
					row[k] = mv
				}
				row[rangeTimestampField] = strconv.FormatFloat(ts, 'f', -1, 64)
				resultName := rangeSeriesResultName(series.Metric, resultNames)
				if resultName != "" {
					row[resultName] = val
				}
				rows = append(rows, row)
			}
		}
	}

	// Inject __ts__ into each stats pipe's ByFields so mergeStatsPipe groups by
	// (by_fields + timestamp) instead of collapsing all timestamps into one row.
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

	// Reconstruct the matrix: group rows by metric key, preserving order.
	type entry struct {
		metric map[string]string
		values []common.ValuePair
	}
	seriesMap := make(map[common.JsonKey]*entry)
	var seriesOrder []common.JsonKey

	for _, row := range rows {
		metricRow := make(map[string]string, len(row))
		tsStr := ""
		valStr := ""

		for k, v := range row {
			if k == rangeTimestampField {
				tsStr = v
			} else if _, isResult := resultNames[k]; isResult {
				valStr = v
			} else {
				metricRow[k] = v
			}
		}

		key, err := common.MakeJsonKey(metricRow)
		if err != nil {
			log.Errorf("stats range: failed to make metric key: %v", err)
			continue
		}
		e, exists := seriesMap[key]
		if !exists {
			e = &entry{metric: metricRow}
			seriesMap[key] = e
			seriesOrder = append(seriesOrder, key)
		}
		ts, err := strconv.ParseFloat(tsStr, 64)
		if err != nil {
			log.Errorf("stats range: failed to parse timestamp %q: %v", tsStr, err)
			continue
		}
		e.values = append(e.values, common.ValuePair{ts, valStr})
	}

	var result StatsRangeResponse
	for _, resp := range responses {
		if resp.Status != "" {
			result.Status = resp.Status
			result.Data.ResultType = resp.Data.ResultType
			break
		}
	}
	result.Data.Result = make([]StatsRangeSeries, 0, len(seriesOrder))
	for _, key := range seriesOrder {
		e := seriesMap[key]
		slices.SortStableFunc(e.values, func(a, b common.ValuePair) int {
			return cmp.Compare(a[0].(float64), b[0].(float64))
		})
		result.Data.Result = append(result.Data.Result, StatsRangeSeries{
			Metric: e.metric,
			Values: e.values,
		})
	}

	return json.Marshal(result)
}

// rangeSeriesResultName returns the result-name field for a series.
// It checks the __name__ metric label first, then falls back to the only
// known result name when there is exactly one.
func rangeSeriesResultName(metric map[string]string, resultNames map[string]struct{}) string {
	if name, ok := metric["__name__"]; ok {
		if _, known := resultNames[name]; known {
			return name
		}
	}
	if len(resultNames) == 1 {
		for rn := range resultNames {
			return rn
		}
	}
	return ""
}
