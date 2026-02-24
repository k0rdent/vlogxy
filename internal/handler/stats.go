package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
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

// Stats handles aggregation of stats query responses from multiple backends
type Stats struct{}

// NewStats creates a new Stats aggregator instance
func NewStats() interfaces.ResponseAggregator[StatsResponse] {
	return &Stats{}
}

func (s *Stats) ParseResponse(resp *http.Response) (StatsResponse, error) {
	return common.DecodeJSONResponse[StatsResponse](resp)
}

func (s *Stats) Merge(responses []StatsResponse) ([]byte, error) {
	groups := make(common.ValuesGroup)

	for _, resp := range responses {
		for _, series := range resp.Data.Result {
			jsonKey, err := common.MakeJsonKey(series.Metric)
			if err != nil {
				log.Errorf("failed to make metric key: %v", err)
				continue
			}

			m, ok := groups[jsonKey]
			if !ok {
				m = make(map[float64]float64)
				groups[jsonKey] = m
			}

			ts, ok := series.Value[0].(float64)
			if !ok {
				log.Errorf("failed to parse timestamp")
				continue
			}

			val, ok := series.Value[1].(string)
			if !ok {
				log.Errorf("failed to parse value")
				continue
			}

			valFloat, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Errorf("failed to parse value as float: %v", err)
				continue
			}
			m[ts] += valFloat
		}
	}

	var result StatsResponse

	// Merge resultType and status from first response
	if len(responses) > 0 {
		result.Data.ResultType = responses[0].Data.ResultType
		result.Status = responses[0].Status
	}

	for jsonKey, tsMap := range groups {
		metric, err := common.ParseJsonKey(jsonKey)
		if err != nil {
			log.Errorf("failed to parse metric key: %v", err)
			continue
		}

		if len(tsMap) == 0 {
			log.Errorf("no timestamps found for metric %v", metric)
			continue
		}
		if len(tsMap) != 1 {
			log.Errorf("expected exactly one timestamp for metric %v, got %d", metric, len(tsMap))
			continue
		}

		var ts float64
		var sum float64
		for t, s := range tsMap {
			ts = t
			sum = s
			break
		}
		valStr := fmt.Sprintf("%f", sum)

		result.Data.Result = append(result.Data.Result, StatsSeries{
			Metric: metric,
			Value:  common.ValuePair{ts, valStr},
		})
	}

	return json.Marshal(result)
}
