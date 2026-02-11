package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/internal/victorialogs"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	HitsArr []Hits `json:"hits"`
}

type Hits struct {
	Fields     map[string]string `json:"fields"`
	Timestamps []string          `json:"timestamps"`
	Values     []int             `json:"values"`
	Total      int               `json:"total"`
}

type HitsQuery struct {
	Path     string
	RawQuery string
}

func NewHits(rawPath, rawQuery string) victorialogs.Querier[Response] {
	return &HitsQuery{
		Path:     rawPath,
		RawQuery: rawQuery,
	}
}

func ProxyHits(c *gin.Context) {
	hits := NewHits(c.Request.URL.Path, c.Request.URL.RawQuery)
	response, err := victorialogs.MultiClusterProxy(hits)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to process hits query",
		})
		return
	}
	c.Data(200, "application/json", response)
}

func (h *HitsQuery) ResponseHandler(resp *http.Response) (Response, error) {
	hit := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(hit); err != nil {
		log.Errorf("failed to decode response: %v", err)
		return Response{}, err
	}
	return *hit, nil
}

func (h *HitsQuery) Merge(responses []Response) ([]byte, error) {
	type group struct {
		fields     map[string]string
		timestamps map[string]int
		total      int
	}

	fieldGroups := make(map[string]*group)

	// Aggregate data from all responses
	for _, resp := range responses {
		for _, hit := range resp.HitsArr {
			// Create a stable key from fields for grouping
			rawFields, err := json.Marshal(hit.Fields)
			if err != nil {
				log.Errorf("failed to marshal fields: %v", err)
				continue
			}
			fieldsKey := string(rawFields)

			// Get or create group
			g, exists := fieldGroups[fieldsKey]
			if !exists {
				g = &group{
					fields:     hit.Fields,
					timestamps: make(map[string]int),
				}
				fieldGroups[fieldsKey] = g
			}

			// Aggregate timestamps and values
			for i, ts := range hit.Timestamps {
				g.timestamps[ts] += hit.Values[i]
			}
			g.total += hit.Total
		}
	}

	// Build final response from aggregated groups
	mergedResponse := &Response{
		HitsArr: make([]Hits, 0, len(fieldGroups)),
	}

	for _, g := range fieldGroups {
		timestamps := make([]string, 0, len(g.timestamps))
		values := make([]int, 0, len(g.timestamps))

		for ts, val := range g.timestamps {
			timestamps = append(timestamps, ts)
			values = append(values, val)
		}

		mergedResponse.HitsArr = append(mergedResponse.HitsArr, Hits{
			Fields:     g.fields,
			Timestamps: timestamps,
			Values:     values,
			Total:      g.total,
		})
	}

	return json.Marshal(mergedResponse)
}

func (h *HitsQuery) GetEndpoint(scheme string, target string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     h.Path,
		RawQuery: h.RawQuery,
	}
	return url.String()
}
