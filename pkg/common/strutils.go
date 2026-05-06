package common

import "strconv"

func FloatToStr(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
