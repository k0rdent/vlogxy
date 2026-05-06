package logstorage

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/timeutil"
)

// TryParseTimestampRFC3339Nano parses s as RFC3339 with optional nanoseconds part and timezone offset and returns unix timestamp in nanoseconds.
//
// If s doesn't contain timezone offset, then the local timezone is used.
//
// The returned timestamp can be negative if s is smaller than 1970 year.
func TryParseTimestampRFC3339Nano(s string) (int64, bool) {
	if len(s) < len("2006-01-02T15:04:05") {
		return 0, false
	}

	secs, ok, tail := tryParseTimestampSecs(s)
	if !ok {
		return 0, false
	}
	s = tail
	nsecs := secs * 1e9

	// Parse timezone offset
	offsetNsecs, prefix, ok := parseTimezoneOffset(s)
	if !ok {
		return 0, false
	}
	nsecs = SubInt64NoOverflow(nsecs, offsetNsecs)
	s = prefix

	// Parse optional fractional part of seconds.
	if len(s) == 0 {
		return nsecs, true
	}
	if s[0] == '.' {
		s = s[1:]
	}
	digits := len(s)
	if digits > 9 {
		return 0, false
	}
	n64, ok := tryParseDateUint64(s)
	if !ok {
		return 0, false
	}

	if digits < 9 {
		n64 *= uint64(math.Pow10(9 - digits))
	}
	nsecs += int64(n64)
	return nsecs, true
}

func parseTimezoneOffset(s string) (int64, string, bool) {
	if strings.HasSuffix(s, "Z") {
		return 0, s[:len(s)-1], true
	}

	n := strings.LastIndexAny(s, "+-")
	if n < 0 {
		offsetNsecs := timeutil.GetLocalTimezoneOffsetNsecs()
		return offsetNsecs, s, true
	}
	offsetStr := s[n+1:]
	isMinus := s[n] == '-'
	if len(offsetStr) == 0 {
		return 0, s, false
	}
	offsetNsecs, ok := tryParseHHMM(offsetStr)
	if !ok {
		return 0, s, false
	}
	if isMinus {
		offsetNsecs = -offsetNsecs
	}
	return offsetNsecs, s[:n], true
}

func tryParseHHMM(s string) (int64, bool) {
	var hourStr, minuteStr string
	switch {
	case len(s) == len("hh:mm") && s[2] == ':':
		hourStr = s[:2]
		minuteStr = s[3:]
	case len(s) == len("hhmm"):
		hourStr = s[:2]
		minuteStr = s[2:]
	default:
		return 0, false
	}
	hours, ok := tryParseDateUint64(hourStr)
	if !ok || hours > 24 {
		return 0, false
	}
	minutes, ok := tryParseDateUint64(minuteStr)
	if !ok || minutes > 60 {
		return 0, false
	}
	return int64(hours)*nsecsPerHour + int64(minutes)*nsecsPerMinute, true
}

// tryParseTimestampSecs parses YYYY-MM-DDTHH:mm:ss into unix timestamp in seconds.
func tryParseTimestampSecs(s string) (int64, bool, string) {
	// Parse year
	if s[len("YYYY")] != '-' {
		return 0, false, s
	}
	yearStr := s[:len("YYYY")]
	n, ok := tryParseDateUint64(yearStr)
	if !ok || n < 1677 || n > 2262 {
		return 0, false, s
	}
	year := int(n)
	s = s[len("YYYY")+1:]

	// Parse month
	if s[len("MM")] != '-' {
		return 0, false, s
	}
	monthStr := s[:len("MM")]
	n, ok = tryParseDateUint64(monthStr)
	if !ok {
		return 0, false, s
	}
	month := time.Month(n)
	s = s[len("MM")+1:]

	// Parse day.
	//
	// Allow whitespace additionally to T as the delimiter after DD,
	// so SQL datetime format can be parsed additionally to RFC3339.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/6721
	delim := s[len("DD")]
	if delim != 'T' && delim != ' ' {
		return 0, false, s
	}
	dayStr := s[:len("DD")]
	n, ok = tryParseDateUint64(dayStr)
	if !ok {
		return 0, false, s
	}
	day := int(n)
	s = s[len("DD")+1:]

	// Parse hour
	if s[len("HH")] != ':' {
		return 0, false, s
	}
	hourStr := s[:len("HH")]
	n, ok = tryParseDateUint64(hourStr)
	if !ok {
		return 0, false, s
	}
	hour := int(n)
	s = s[len("HH")+1:]

	// Parse minute
	if s[len("MM")] != ':' {
		return 0, false, s
	}
	minuteStr := s[:len("MM")]
	n, ok = tryParseDateUint64(minuteStr)
	if !ok {
		return 0, false, s
	}
	minute := int(n)
	s = s[len("MM")+1:]

	// Parse second
	secondStr := s[:len("SS")]
	n, ok = tryParseDateUint64(secondStr)
	if !ok {
		return 0, false, s
	}
	second := int(n)
	s = s[len("SS"):]

	secs := time.Date(year, month, day, hour, minute, second, 0, time.UTC).Unix()
	if secs < int64(-1<<63)/1e9 || secs >= int64((1<<63)-1)/1e9 {
		// Too big or too small timestamp
		return 0, false, s
	}
	return secs, true, s
}

// tryParseUint64 parses s as uint64 value.
func tryParseUint64(s string) (uint64, bool) {
	if len(s) == 0 || len(s) > len("18_446_744_073_709_551_615") {
		return 0, false
	}
	if len(s) > 1 && s[0] == '0' {
		// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/8361
		return 0, false
	}

	n := uint64(0)
	for i := range len(s) {
		ch := s[i]
		if ch == '_' {
			continue
		}
		if ch < '0' || ch > '9' {
			return 0, false
		}
		if n > ((1<<64)-1)/10 {
			// overflow
			return 0, false
		}
		n *= 10
		d := uint64(ch - '0')
		n1 := n + d
		if n1 < n {
			// overflow
			return 0, false
		}
		n = n1
	}
	return n, true
}

// tryParseDateUint64 parses s (which is a part of some timestamp) as uint64 value.
func tryParseDateUint64(s string) (uint64, bool) {
	if len(s) == 0 || len(s) > 9 {
		return 0, false
	}

	if len(s) == 2 {
		// fast path for two-digit number, which is used in hours, minutes and seconds
		if s[0] < '0' || s[0] > '9' {
			return 0, false
		}
		n := 10*uint64(s[0]-'0') + uint64(s[1]-'0')
		return n, true
	}

	n := uint64(0)
	for i := range len(s) {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, false
		}
		if n > ((1<<64)-1)/10 {
			return 0, false
		}
		n *= 10
		d := uint64(ch - '0')
		if n > (1<<64)-1-d {
			return 0, false
		}
		n += d
	}
	return n, true
}

// tryParseIPv4 tries parsing ipv4 from s.
func tryParseIPv4(s string) (uint32, bool) {
	if len(s) < len("1.1.1.1") || len(s) > len("255.255.255.255") || strings.Count(s, ".") != 3 {
		// Fast path - the entry isn't IPv4
		return 0, false
	}

	var octets [4]byte
	var v uint64
	var ok bool

	// Parse octet 1
	n := strings.IndexByte(s, '.')
	if n <= 0 || n > 3 {
		return 0, false
	}
	v, ok = tryParseDateUint64(s[:n])
	if !ok || v > 255 {
		return 0, false
	}
	octets[0] = byte(v)
	s = s[n+1:]

	// Parse octet 2
	n = strings.IndexByte(s, '.')
	if n <= 0 || n > 3 {
		return 0, false
	}
	v, ok = tryParseDateUint64(s[:n])
	if !ok || v > 255 {
		return 0, false
	}
	octets[1] = byte(v)
	s = s[n+1:]

	// Parse octet 3
	n = strings.IndexByte(s, '.')
	if n <= 0 || n > 3 {
		return 0, false
	}
	v, ok = tryParseDateUint64(s[:n])
	if !ok || v > 255 {
		return 0, false
	}
	octets[2] = byte(v)
	s = s[n+1:]

	// Parse octet 4
	v, ok = tryParseDateUint64(s)
	if !ok || v > 255 {
		return 0, false
	}
	octets[3] = byte(v)

	ipv4 := encoding.UnmarshalUint32(octets[:])
	return ipv4, true
}

// tryParseFloat64Prefix tries parsing float64 number at the beginning of s and returns the remaining tail.
func tryParseFloat64Prefix(s string) (float64, bool, string) {
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.' || s[i] == '_') {
		i++
	}
	if i == 0 {
		return 0, false, s
	}

	f, ok := tryParseFloat64(s[:i])
	return f, ok, s[i:]
}

// tryParseFloat64 tries parsing s as float64.
//
// The parsed result may lose precision, e.g. it may not match the original value when converting back to string.
// Use tryParseFloat64Exact when lossless parsing is needed.
func tryParseFloat64(s string) (float64, bool) {
	return tryParseFloat64Internal(s, false)
}

func tryParseFloat64Internal(s string, isExact bool) (float64, bool) {
	if len(s) == 0 || len(s) > len("-18_446_744_073_709_551_615") {
		return 0, false
	}
	// Allow only decimal digits, minus and a dot.
	// Do not allows scientific notation (for example 1.23E+05),
	// since it cannot be converted back to the same string form.

	minus := s[0] == '-'
	if minus {
		s = s[1:]
	}
	n := strings.IndexByte(s, '.')
	if n < 0 {
		// fast path - there are no dots
		n, ok := tryParseUint64(s)
		if !ok {
			return 0, false
		}
		if isExact && n >= (1<<53) {
			// The integer cannot be represented as float64 without precision loss.
			return 0, false
		}
		f := float64(n)
		if minus {
			f = -f
		}
		return f, true
	}
	if n == 0 || n == len(s)-1 {
		// Do not allow dots at the beginning and at the end of s,
		// since they cannot be converted back to the same string form.
		return 0, false
	}
	sInt := s[:n]
	sFrac := s[n+1:]

	nInt, ok := tryParseUint64(sInt)
	if !ok {
		return 0, false
	}

	// Skip leading zeroes at sFrac, since tryParseUint64 rejects them.
	// This fixes https://github.com/VictoriaMetrics/VictoriaMetrics/issues/8464
	n = 0
	for n < len(sFrac)-1 && sFrac[n] == '0' {
		n++
	}

	nFrac, ok := tryParseUint64(sFrac[n:])
	if !ok {
		return 0, false
	}

	p10 := math.Pow10(strings.Count(sFrac, "_") - len(sFrac))
	f := math.FMA(float64(nFrac), p10, float64(nInt))
	if minus {
		f = -f
	}
	return f, true
}

// tryParseBytes parses user-readable bytes representation in s.
//
// Supported suffixes:
//
//	K, KB - for 1000
func tryParseBytes(s string) (int64, bool) {
	if len(s) == 0 {
		return 0, false
	}

	isMinus := s[0] == '-'
	if isMinus {
		s = s[1:]
	}

	n := int64(0)
	for len(s) > 0 {
		f, ok, tail := tryParseFloat64Prefix(s)
		if !ok {
			return 0, false
		}
		if len(tail) == 0 {
			if _, frac := math.Modf(f); frac != 0 {
				// deny floating-point numbers without any suffix.
				return 0, false
			}
		}
		s = tail
		if len(s) == 0 {
			n = addInt64NoOverflow(n, f)
			continue
		}
		if len(s) >= 3 {
			switch {
			case strings.HasPrefix(s, "KiB"):
				n = addInt64NoOverflow(n, f*(1<<10))
				s = s[3:]
				continue
			case strings.HasPrefix(s, "MiB"):
				n = addInt64NoOverflow(n, f*(1<<20))
				s = s[3:]
				continue
			case strings.HasPrefix(s, "GiB"):
				n = addInt64NoOverflow(n, f*(1<<30))
				s = s[3:]
				continue
			case strings.HasPrefix(s, "TiB"):
				n = addInt64NoOverflow(n, f*(1<<40))
				s = s[3:]
				continue
			}
		}
		if len(s) >= 2 {
			switch {
			case strings.HasPrefix(s, "Ki"):
				n = addInt64NoOverflow(n, f*(1<<10))
				s = s[2:]
				continue
			case strings.HasPrefix(s, "Mi"):
				n = addInt64NoOverflow(n, f*(1<<20))
				s = s[2:]
				continue
			case strings.HasPrefix(s, "Gi"):
				n = addInt64NoOverflow(n, f*(1<<30))
				s = s[2:]
				continue
			case strings.HasPrefix(s, "Ti"):
				n = addInt64NoOverflow(n, f*(1<<40))
				s = s[2:]
				continue
			case strings.HasPrefix(s, "KB"):
				n = addInt64NoOverflow(n, f*1_000)
				s = s[2:]
				continue
			case strings.HasPrefix(s, "MB"):
				n = addInt64NoOverflow(n, f*1_000_000)
				s = s[2:]
				continue
			case strings.HasPrefix(s, "GB"):
				n = addInt64NoOverflow(n, f*1_000_000_000)
				s = s[2:]
				continue
			case strings.HasPrefix(s, "TB"):
				n = addInt64NoOverflow(n, f*1_000_000_000_000)
				s = s[2:]
				continue
			}
		}
		switch {
		case strings.HasPrefix(s, "B"):
			n = addInt64NoOverflow(n, f)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "K"):
			n = addInt64NoOverflow(n, f*1_000)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "M"):
			n = addInt64NoOverflow(n, f*1_000_000)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "G"):
			n = addInt64NoOverflow(n, f*1_000_000_000)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "T"):
			n = addInt64NoOverflow(n, f*1_000_000_000_000)
			s = s[1:]
			continue
		}
	}

	if isMinus {
		n = -n
	}
	return n, true
}

func addInt64NoOverflow(n int64, f float64) int64 {
	x := int64(f)
	if n < 0 || x < 0 || x > 1<<63-1-n {
		return 1<<63 - 1
	}
	return n + x
}

// tryParseIPv4Mask parses '/num' ipv4 mask and returns (1<<(32-num))
func tryParseIPv4Mask(s string) (uint64, bool) {
	if len(s) == 0 || s[0] != '/' {
		return 0, false
	}
	s = s[1:]
	n, ok := tryParseUint64(s)
	if !ok || n > 32 {
		return 0, false
	}
	return 1 << (32 - uint8(n)), true
}

// tryParseDuration parses the given duration in nanoseconds and returns the result.
func tryParseDuration(s string) (int64, bool) {
	if len(s) == 0 {
		return 0, false
	}
	isMinus := s[0] == '-'
	if isMinus {
		s = s[1:]
	}

	nsecs := int64(0)
	for len(s) > 0 {
		f, ok, tail := tryParseFloat64Prefix(s)
		if !ok {
			return 0, false
		}
		s = tail
		if len(s) == 0 {
			return 0, false
		}
		if len(s) >= 3 {
			if strings.HasPrefix(s, "µs") {
				nsecs = addInt64NoOverflow(nsecs, f*nsecsPerMicrosecond)
				s = s[3:]
				continue
			}
		}
		if len(s) >= 2 {
			switch {
			case strings.HasPrefix(s, "ms"):
				nsecs = addInt64NoOverflow(nsecs, f*nsecsPerMillisecond)
				s = s[2:]
				continue
			case strings.HasPrefix(s, "ns"):
				nsecs = addInt64NoOverflow(nsecs, f)
				s = s[2:]
				continue
			}
		}
		switch {
		case strings.HasPrefix(s, "y"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerYear)
			s = s[1:]
		case strings.HasPrefix(s, "w"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerWeek)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "d"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerDay)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "h"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerHour)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "m"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerMinute)
			s = s[1:]
			continue
		case strings.HasPrefix(s, "s"):
			nsecs = addInt64NoOverflow(nsecs, f*nsecsPerSecond)
			s = s[1:]
			continue
		default:
			return 0, false
		}
	}

	if isMinus {
		nsecs = -nsecs
	}
	return nsecs, true
}

// marshalDurationString appends string representation of nsec duration to dst and returns the result.
func marshalDurationString(dst []byte, nsecs int64) []byte {
	if nsecs == 0 {
		return append(dst, '0')
	}

	if nsecs < 0 {
		dst = append(dst, '-')
		nsecs = -nsecs
	}
	formatFloat64Seconds := nsecs >= nsecsPerSecond

	if nsecs >= nsecsPerWeek {
		weeks := nsecs / nsecsPerWeek
		nsecs -= weeks * nsecsPerWeek
		dst = marshalUint64String(dst, uint64(weeks))
		dst = append(dst, 'w')
	}
	if nsecs >= nsecsPerDay {
		days := nsecs / nsecsPerDay
		nsecs -= days * nsecsPerDay
		dst = marshalUint8String(dst, uint8(days))
		dst = append(dst, 'd')
	}
	if nsecs >= nsecsPerHour {
		hours := nsecs / nsecsPerHour
		nsecs -= hours * nsecsPerHour
		dst = marshalUint8String(dst, uint8(hours))
		dst = append(dst, 'h')
	}
	if nsecs >= nsecsPerMinute {
		minutes := nsecs / nsecsPerMinute
		nsecs -= minutes * nsecsPerMinute
		dst = marshalUint8String(dst, uint8(minutes))
		dst = append(dst, 'm')
	}
	if nsecs >= nsecsPerSecond {
		if formatFloat64Seconds {
			seconds := float64(nsecs) / nsecsPerSecond
			dst = marshalFloat64String(dst, seconds)
			dst = append(dst, 's')
			return dst
		}
		seconds := nsecs / nsecsPerSecond
		nsecs -= seconds * nsecsPerSecond
		dst = marshalUint8String(dst, uint8(seconds))
		dst = append(dst, 's')
	}
	if nsecs >= nsecsPerMillisecond {
		msecs := nsecs / nsecsPerMillisecond
		nsecs -= msecs * nsecsPerMillisecond
		dst = marshalUint16String(dst, uint16(msecs))
		dst = append(dst, "ms"...)
	}
	if nsecs >= nsecsPerMicrosecond {
		usecs := nsecs / nsecsPerMicrosecond
		nsecs -= usecs * nsecsPerMicrosecond
		dst = marshalUint16String(dst, uint16(usecs))
		dst = append(dst, "µs"...)
	}
	if nsecs > 0 {
		dst = marshalUint16String(dst, uint16(nsecs))
		dst = append(dst, "ns"...)
	}
	return dst
}

const (
	nsecsPerYear        = 365 * 24 * 3600 * 1e9
	nsecsPerWeek        = 7 * 24 * 3600 * 1e9
	nsecsPerDay         = 24 * 3600 * 1e9
	nsecsPerHour        = 3600 * 1e9
	nsecsPerMinute      = 60 * 1e9
	nsecsPerSecond      = 1e9
	nsecsPerMillisecond = 1e6
	nsecsPerMicrosecond = 1e3
)

func marshalUint8String(dst []byte, n uint8) []byte {
	if n < 10 {
		return append(dst, '0'+n)
	}
	if n < 100 {
		return append(dst, '0'+n/10, '0'+n%10)
	}

	if n < 200 {
		dst = append(dst, '1')
		n -= 100
	} else {
		dst = append(dst, '2')
		n -= 200
	}
	if n < 10 {
		return append(dst, '0', '0'+n)
	}
	return append(dst, '0'+n/10, '0'+n%10)
}

func marshalUint16String(dst []byte, n uint16) []byte {
	return marshalUint64String(dst, uint64(n))
}

func marshalUint64String(dst []byte, n uint64) []byte {
	return strconv.AppendUint(dst, n, 10)
}

func marshalFloat64String(dst []byte, f float64) []byte {
	return strconv.AppendFloat(dst, f, 'f', -1, 64)
}

func marshalIPv4String(dst []byte, n uint32) []byte {
	dst = marshalUint8String(dst, uint8(n>>24))
	dst = append(dst, '.')
	dst = marshalUint8String(dst, uint8(n>>16))
	dst = append(dst, '.')
	dst = marshalUint8String(dst, uint8(n>>8))
	dst = append(dst, '.')
	dst = marshalUint8String(dst, uint8(n))
	return dst
}
