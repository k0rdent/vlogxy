package handler

import (
	"encoding/json"
	"net/http"

	"github.com/k0rdent/victorialogs-aggregator/internal/interfaces"
	"github.com/k0rdent/victorialogs-aggregator/pkg/common"
)

type FieldValuesResponse struct {
	Values []Value `json:"values"`
}

type Value struct {
	Value string `json:"value"`
	Hits  int    `json:"hits"`
}

type FieldValuesQuery struct {
	*common.RequestPath
}

func NewFieldValuesQuery(path, rawQuery string) interfaces.ResponseAggregator[FieldValuesResponse] {
	return &FieldValuesQuery{
		RequestPath: &common.RequestPath{
			Path:     path,
			RawQuery: rawQuery,
		},
	}
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

func (f *FieldValuesQuery) GetURL(scheme string, target string) string {
	return common.BuildURL(scheme, target, f.Path, f.RawQuery)
}
