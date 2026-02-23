package handler

import (
	"encoding/json"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
)

// ValuesResponse represents the shared response structure for any endpoint
// that returns a flat list of values with hit counts.
type ValuesResponse struct {
	Values []Value `json:"values"`
}

// Value represents a single field value with its hit count.
type Value struct {
	Value string `json:"value"`
	Hits  int    `json:"hits"`
}

// ValueHitsAggregator is a generic aggregator for all endpoints that return
// a ValuesResponse. The same struct is reused across field_values,
// field_names, streams, stream_ids, stream_field_values, and stream_field_names.
type ValueHitsAggregator struct{}

func (a *ValueHitsAggregator) ParseResponse(resp *http.Response) (ValuesResponse, error) {
	return common.DecodeJSONResponse[ValuesResponse](resp)
}

func (a *ValueHitsAggregator) Merge(responses []ValuesResponse) ([]byte, error) {
	return json.Marshal(mergeValueHits(responses))
}

// mergeValueHits aggregates hit counts for identical values across multiple
// ValuesResponse instances and returns a single combined response.
func mergeValueHits(responses []ValuesResponse) ValuesResponse {
	valueMap := make(map[string]int)

	for _, resp := range responses {
		for _, v := range resp.Values {
			valueMap[v.Value] += v.Hits
		}
	}

	result := ValuesResponse{Values: make([]Value, 0, len(valueMap))}
	for val, hits := range valueMap {
		result.Values = append(result.Values, Value{Value: val, Hits: hits})
	}

	return result
}

// newValueHitsAggregator returns a ValueHitsAggregator satisfying ResponseAggregator[ValuesResponse].
func newValueHitsAggregator() interfaces.ResponseAggregator[ValuesResponse] {
	return &ValueHitsAggregator{}
}

func NewFieldValuesQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
func NewFieldNamesQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
func NewStreamsQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
func NewStreamIdsQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
func NewStreamFieldValuesQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
func NewStreamFieldNamesQuery() interfaces.ResponseAggregator[ValuesResponse] {
	return newValueHitsAggregator()
}
