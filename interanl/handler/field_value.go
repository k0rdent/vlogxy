package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/interanl/victorialogs"
	log "github.com/sirupsen/logrus"
)

type FieldValuesResponse struct {
	Values []Value `json:"values"`
}

type Value struct {
	Value string `json:"value"`
	Hits  int    `json:"hits"`
}

type FieldValuesQuery struct {
	Path     string
	RawQuery string
}

func NewFieldValuesQuery(rawPath, rawQuery string) victorialogs.Querier[FieldValuesResponse] {
	return &FieldValuesQuery{
		Path:     rawPath,
		RawQuery: rawQuery,
	}
}

func ProxyFieldValues(c *gin.Context) {
	query := NewFieldValuesQuery(c.Request.URL.Path, c.Request.URL.RawQuery)
	response, err := victorialogs.MultiClusterProxy(query)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to process field values query",
		})
		return
	}
	c.Data(200, "application/json", response)
}

func (f *FieldValuesQuery) ResponseHandler(resp *http.Response) (FieldValuesResponse, error) {
	response := new(FieldValuesResponse)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		log.Errorf("failed to decode response: %v", err)
		return FieldValuesResponse{}, err
	}
	return *response, nil
}

func (f *FieldValuesQuery) Merge(responses []FieldValuesResponse) ([]byte, error) {
	mergedResponse := new(FieldValuesResponse)
	valueMap := make(map[string]int)

	for _, resp := range responses {
		for _, value := range resp.Values {
			if val, exists := valueMap[value.Value]; exists {
				valueMap[value.Value] = val + value.Hits
			} else {
				valueMap[value.Value] = value.Hits
			}
		}
	}

	for val, hits := range valueMap {
		mergedResponse.Values = append(mergedResponse.Values, Value{
			Value: val,
			Hits:  hits,
		})
	}

	return json.Marshal(mergedResponse)
}

func (f *FieldValuesQuery) GetEndpoint(scheme string, target string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     f.Path,
		RawQuery: f.RawQuery,
	}
	return url.String()
}
