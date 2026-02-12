package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

type StatsRangeResponse struct {
	Data struct {
		ResultType string             `json:"resultType"`
		Result     []StatsRangeSeries `json:"result"`
	} `json:"data"`
	Status string `json:"status"`
}

type StatsRangeSeries struct {
	Metric map[string]string  `json:"metric"`
	Values []common.ValuePair `json:"values"`
}

type StatsRange struct {
	*common.RequestPath
}

func NewStatsRange(path, rawQuery string) interfaces.ResponseAggregator[StatsRangeResponse] {
	return &StatsRange{
		RequestPath: &common.RequestPath{
			Path:     path,
			RawQuery: rawQuery,
		},
	}
}

func (s *StatsRange) ParseResponse(resp *http.Response) (StatsRangeResponse, error) {
	return common.DecodeJSONResponse[StatsRangeResponse](resp)
}

func (s *StatsRange) Merge(responses []StatsRangeResponse) ([]byte, error) {
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
				m = make(map[float64]int64)
				groups[jsonKey] = m
			}

			for _, v := range series.Values {
				ts, ok := v[0].(float64)
				if !ok {
					log.Errorf("failed to parse timestamp")
					continue
				}

				val, ok := v[1].(string)
				if !ok {
					log.Errorf("failed to parse value")
					continue
				}

				valInt, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					log.Errorf("failed to parse value as int: %v", err)
					continue
				}
				m[ts] += valInt
			}
		}
	}

	var result StatsRangeResponse

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

		var values []common.ValuePair
		for ts := range tsMap {
			valStr := fmt.Sprintf("%d", tsMap[ts])
			values = append(values, common.ValuePair{ts, valStr})
		}

		slices.SortFunc(values, func(a, b common.ValuePair) int {
			tsA := a[0].(float64)
			tsB := b[0].(float64)
			if tsA < tsB {
				return -1
			} else if tsA > tsB {
				return 1
			}
			return 0
		})

		result.Data.Result = append(result.Data.Result, StatsRangeSeries{
			Metric: metric,
			Values: values,
		})
	}

	return json.Marshal(result)
}

func (s *StatsRange) GetURL(scheme, target, path string) string {
	return common.BuildURL(scheme, target, path+s.Path, s.RawQuery)
}
