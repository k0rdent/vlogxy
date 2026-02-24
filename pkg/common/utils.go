package common

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// MakeJsonKey creates a stable key from metric labels for grouping
func MakeJsonKey(m map[string]string) (JsonKey, error) {
	rawKey, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return JsonKey(rawKey), nil
}

// ParseJsonKey parses a JSON key back into labels map
func ParseJsonKey(k JsonKey) (map[string]string, error) {
	var m map[string]string
	if err := json.Unmarshal([]byte(k), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// DecodeJSONResponse is a generic helper for decoding JSON responses
func DecodeJSONResponse[T any](resp *http.Response) (T, error) {
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// BuildURL constructs a full URL from scheme, host, path and query parameters
func BuildURL(scheme, host, path, rawQuery string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: rawQuery,
	}
	return url.String()
}

// GetOrCreateInnerMap retrieves an inner map from an outer map, creating it if it doesn't exist
func GetOrCreateInnerMap[K comparable, V ~map[string]int](outerMap map[K]V, key K) V {
	innerMap, ok := outerMap[key]
	if !ok {
		innerMap = make(V)
		outerMap[key] = innerMap
	}
	return innerMap
}
