package handler

import (
	"encoding/json"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
)

type FieldValuesResponse struct {
	Values []Value `json:"values"`
}

type Value struct {
	Value string `json:"value"`
	Hits  int    `json:"hits"`
}

type FieldValuesQuery struct{}

func NewFieldValuesQuery() interfaces.ResponseAggregator[FieldValuesResponse] {
	return &FieldValuesQuery{}
}

func (f *FieldValuesQuery) ParseResponse(resp *http.Response) (FieldValuesResponse, error) {
	return common.DecodeJSONResponse[FieldValuesResponse](resp)
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
