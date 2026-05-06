package aggregation

import (
	"strconv"

	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

type AggFunc int

const (
	// AggSum sums values across backends (e.g. count, sum, count_uniq, count_empty, sum_len, rate, rate_sum).
	AggSum AggFunc = iota
	// AggMin takes the minimum value across backends (e.g. min, first).
	AggMin
	// AggMax takes the maximum value across backends (e.g. max, last).
	AggMax
	// AggAvg averages values across backends (e.g. avg, median, quantile).
	AggAvg
)

const (
	FuncBy         = "by"
	FuncCount      = "count"
	FuncCountUniq  = "count_uniq"
	FuncCountEmpty = "count_empty"
	FuncSum        = "sum"
	FuncSumLen     = "sum_len"
	FuncRate       = "rate"
	FuncRateSum    = "rate_sum"
	FuncMin        = "min"
	FuncFirst      = "first"
	FuncMax        = "max"
	FuncLast       = "last"
	FuncAvg        = "avg"
	FuncMedian     = "median"
	FuncQuantile   = "quantile"
)

// funcToAgg maps lowercase VictoriaLogs stats function names to their aggregation strategy.
var funcToAgg = map[string]AggFunc{
	FuncCount:      AggSum,
	FuncCountUniq:  AggSum,
	FuncCountEmpty: AggSum,
	FuncSum:        AggSum,
	FuncSumLen:     AggSum,
	FuncRate:       AggSum,
	FuncRateSum:    AggSum,
	FuncMin:        AggMin,
	FuncFirst:      AggMin,
	FuncMax:        AggMax,
	FuncLast:       AggMax,
	FuncAvg:        AggAvg,
	FuncMedian:     AggAvg,
	FuncQuantile:   AggAvg,
}

// Aggregate takes a function name and a list of string values from different backends, and merges them according to the aggregation strategy for that function. The merged result is returned as a string.
func Aggregate(funcName string, values []string) string {
	aggFunc, ok := funcToAgg[funcName]
	if !ok {
		// If the function is not recognized, return the first value (arbitrary choice).
		return values[0]
	}

	switch aggFunc {
	case AggSum:
		return AggregateSum(values)
	case AggMin:
		return AggregateMin(values)
	case AggMax:
		return AggregateMax(values)
	case AggAvg:
		return AggregateAvg(values)
	default:
		return values[0]
	}
}

// AggregateSum sums values across backends (e.g. count, sum, count_uniq, count_empty, sum_len, rate, rate_sum).
func AggregateSum(values []string) string {
	var sum float64
	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Errorf("failed to parse value %q as float for sum aggregation: %v", v, err)
			continue
		}
		sum += f
	}
	return common.FloatToStr(sum)
}

// AggregateMin takes the minimum value across backends (e.g. min, first).
func AggregateMin(values []string) string {
	var min float64
	for i, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Errorf("failed to parse value %q as float for min aggregation: %v", v, err)
			continue
		}
		if i == 0 || f < min {
			min = f
		}
	}
	return common.FloatToStr(min)
}

// AggregateMax takes the maximum value across backends (e.g. max, last).
func AggregateMax(values []string) string {
	var max float64
	for i, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Errorf("failed to parse value %q as float for max aggregation: %v", v, err)
			continue
		}
		if i == 0 || f > max {
			max = f
		}
	}
	return common.FloatToStr(max)
}

// AggregateAvg averages values across backends (e.g. avg, median, quantile).
func AggregateAvg(values []string) string {
	var sum float64
	var count int
	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Errorf("failed to parse value %q as float for avg aggregation: %v", v, err)
			continue
		}
		sum += f
		count++
	}
	if count == 0 {
		return "0"
	}
	avg := sum / float64(count)
	return common.FloatToStr(avg)
}

func GetAggFunc(funcName string) AggFunc {
	return funcToAgg[funcName]
}

func IsAggFunc(funcName string) bool {
	_, ok := funcToAgg[funcName]
	return ok
}
