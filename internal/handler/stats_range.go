package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/internal/victorialogs"
	log "github.com/sirupsen/logrus"
)

type StatsRange struct {
	Path     string
	RawQuery string
}

type StatsRangeResponse struct {
	Data struct {
		ResultType string             `json:"resultType"`
		Result     []StatsRangeSeries `json:"result"`
	} `json:"data"`
	Status string `json:"status"`
}

type StatsRangeSeries struct {
	Metric map[string]string `json:"metric"`
	// Values is an array of [timestamp, value] pairs
	// where timestamp is a float64 and value is a string
	Values []ValuePair `json:"values"`
}

func ProxyStatsRange(c *gin.Context) {
	query := NewStatsRange(c.Request.URL.Path, c.Request.URL.RawQuery)
	response, err := victorialogs.MultiClusterProxy(query)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to process stats range query",
		})
		return
	}
	c.Data(200, "application/json", response)
}

func NewStatsRange(rawPath, rawQuery string) victorialogs.Querier[StatsRangeResponse] {
	return &StatsRange{
		Path:     rawPath,
		RawQuery: rawQuery,
	}
}

func (s *StatsRange) ResponseHandler(resp *http.Response) (StatsRangeResponse, error) {
	statsResp := new(StatsRangeResponse)
	if err := json.NewDecoder(resp.Body).Decode(statsResp); err != nil {
		log.Errorf("failed to decode response: %v", err)
		return StatsRangeResponse{}, err
	}
	return *statsResp, nil
}

type key string
type ValuesGroup map[key]map[float64]int64 // metricKey -> timestamp -> sum

func (s *StatsRange) Merge(responses []StatsRangeResponse) ([]byte, error) {
	groups := make(ValuesGroup)
	for _, resp := range responses {
		for _, series := range resp.Data.Result {
			metricKey, err := makeMetricKey(series.Metric)
			if err != nil {
				log.Errorf("failed to make metric key: %v", err)
				continue
			}

			m, ok := groups[metricKey]
			if !ok {
				m = make(map[float64]int64)
				groups[metricKey] = m
			}

			for _, v := range series.Values {
				ts, ok := v[0].(float64)
				if !ok {
					log.Errorf("failed to parse timestamp: %v", err)
					continue
				}

				val, ok := v[1].(string)
				if !ok {
					log.Errorf("failed to parse value: %v", err)
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

	for metricKey, tsMap := range groups {
		metric, err := parseMetricKey(metricKey)
		if err != nil {
			log.Errorf("failed to parse metric key: %v", err)
			continue
		}
		var values []ValuePair
		for ts := range tsMap {
			valStr := fmt.Sprintf("%d", tsMap[ts])
			values = append(values, ValuePair{ts, valStr})
		}

		result.Data.Result = append(result.Data.Result, StatsRangeSeries{
			Metric: metric,
			Values: values,
		})
	}

	return json.Marshal(result)
}

func (s *StatsRange) GetEndpoint(scheme string, target string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     s.Path,
		RawQuery: s.RawQuery,
	}
	return url.String()
}

func makeMetricKey(m map[string]string) (key, error) {
	rawKey, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return key(rawKey), nil
}

func parseMetricKey(k key) (map[string]string, error) {
	var m map[string]string
	if err := json.Unmarshal([]byte(k), &m); err != nil {
		return nil, err
	}
	return m, nil
}
