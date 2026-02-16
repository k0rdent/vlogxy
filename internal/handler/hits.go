package handler

import (
	"encoding/json"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

// Response represents the response structure for hits queries
type Response struct {
	HitsArr []Hits `json:"hits"`
}

// Hits represents aggregated hit data for a specific field combination
type Hits struct {
	Fields     map[string]string `json:"fields"`
	Timestamps []string          `json:"timestamps"`
	Values     []int             `json:"values"`
	Total      int               `json:"total"`
}

// HitsQuery handles aggregation of hits query responses from multiple backends
type HitsQuery struct{}

// NewHits creates a new HitsQuery aggregator instance
func NewHits() interfaces.ResponseAggregator[Response] {
	return &HitsQuery{}
}

func (h *HitsQuery) ParseResponse(resp *http.Response) (Response, error) {
	return common.DecodeJSONResponse[Response](resp)
}

func (h *HitsQuery) Merge(responses []Response) ([]byte, error) {
	type group struct {
		fields     map[string]string
		timestamps map[string]int
		total      int
	}

	fieldGroups := make(map[common.JsonKey]*group)

	for _, resp := range responses {
		for _, hit := range resp.HitsArr {
			key, err := common.MakeJsonKey(hit.Fields)
			if err != nil {
				log.Errorf("failed to make key: %v", err)
				continue
			}

			g, exists := fieldGroups[key]
			if !exists {
				g = &group{
					fields:     hit.Fields,
					timestamps: make(map[string]int),
				}
				fieldGroups[key] = g
			}

			for i, ts := range hit.Timestamps {
				g.timestamps[ts] += hit.Values[i]
			}
			g.total += hit.Total
		}
	}

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
			Timestamps: timestamps,
			Values:     values,
			Total:      g.total,
			Fields:     g.fields,
		})
	}

	return json.Marshal(mergedResponse)
}
