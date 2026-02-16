package common

// ValuePair represents a single [timestamp, value] pair
// where timestamp is a float64 and value is a string
type ValuePair [2]any

// JsonKey is a unique identifier for a metric based on its labels
type JsonKey string

// ValuesGroup maps metric keys to their aggregated values over time
// Used for merging time series data from multiple backends
type ValuesGroup map[JsonKey]map[float64]int64
