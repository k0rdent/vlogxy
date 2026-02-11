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

type Stats struct {
	Path     string
	RawQuery string
}

type StatsResponse struct {
	Data struct {
		ResultType string        `json:"resultType"`
		Result     []StatsSeries `json:"result"`
	} `json:"data"`
	Status string `json:"status"`
}

type StatsSeries struct {
	Metric map[string]string `json:"metric"`
	Value  ValuePair         `json:"value"`
}

// ValuePair is a single [timestamp, value] pair
// where timestamp is a float64 and value is a string
type ValuePair [2]any

func ProxyStats(c *gin.Context) {
	query := NewStats(c.Request.URL.Path, c.Request.URL.RawQuery)
	response, err := victorialogs.MultiClusterProxy(query)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to process stats query",
		})
		return
	}
	c.Data(200, "application/json", response)
}

func NewStats(rawPath, rawQuery string) victorialogs.Querier[StatsResponse] {
	return &Stats{
		Path:     rawPath,
		RawQuery: rawQuery,
	}
}

func (s *Stats) ResponseHandler(resp *http.Response) (StatsResponse, error) {
	statsResp := new(StatsResponse)
	if err := json.NewDecoder(resp.Body).Decode(statsResp); err != nil {
		log.Errorf("failed to decode response: %v", err)
		return StatsResponse{}, err
	}
	return *statsResp, nil
}

func (s *Stats) Merge(responses []StatsResponse) ([]byte, error) {
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

			ts, ok := series.Value[0].(float64)
			if !ok {
				log.Errorf("failed to parse timestamp: %v", err)
				continue
			}

			val, ok := series.Value[1].(string)
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

	var result StatsResponse

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
		// For aggregated result, we take the first timestamp we have
		var ts float64
		var sum int64
		for t, s := range tsMap {
			ts = t
			sum = s
			break // Only one timestamp per metric in the response
		}
		valStr := fmt.Sprintf("%d", sum)

		result.Data.Result = append(result.Data.Result, StatsSeries{
			Metric: metric,
			Value:  ValuePair{ts, valStr},
		})
	}

	return json.Marshal(result)
}

func (s *Stats) GetEndpoint(scheme string, target string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     s.Path,
		RawQuery: s.RawQuery,
	}
	return url.String()
}
