package handler

import (
	"encoding/json"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
)

type FacetsResponse struct {
	Facets []Facets `json:"facets"`
}

type Facets struct {
	FieldName string       `json:"field_name"`
	Values    []FacetValue `json:"values"`
}

type FacetValue struct {
	Value string `json:"field_value"`
	Hits  int    `json:"hits"`
}

type FacetsMap map[string]FacetsValuesMap
type FacetsValuesMap map[string]int

type FacetsQuery struct{}

func NewFacetsQuery() interfaces.ResponseAggregator[FacetsResponse] {
	return &FacetsQuery{}
}

func (f *FacetsQuery) ParseResponse(resp *http.Response) (FacetsResponse, error) {
	return common.DecodeJSONResponse[FacetsResponse](resp)
}

func (f *FacetsQuery) Merge(responses []FacetsResponse) ([]byte, error) {
	mergedResponse := new(FacetsResponse)
	facetsMap := make(FacetsMap)

	for _, resp := range responses {
		for _, facet := range resp.Facets {
			facetValuesMap := common.GetOrCreateInnerMap(facetsMap, facet.FieldName)

			for _, value := range facet.Values {
				facetValuesMap[value.Value] += value.Hits
			}
		}
	}

	mergedResponse.Facets = make([]Facets, 0, len(facetsMap))
	for fieldName, valuesMap := range facetsMap {
		facet := Facets{
			FieldName: fieldName,
			Values:    make([]FacetValue, 0, len(valuesMap)),
		}
		for value, hits := range valuesMap {
			facet.Values = append(facet.Values, FacetValue{
				Value: value,
				Hits:  hits,
			})
		}
		mergedResponse.Facets = append(mergedResponse.Facets, facet)
	}

	return json.Marshal(mergedResponse)
}
