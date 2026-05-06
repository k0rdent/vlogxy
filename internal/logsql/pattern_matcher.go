package logstorage

import (
	"strings"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

type patternMatcher struct {
	pmo patternMatcherOption

	separators   []string
	placeholders []patternMatcherPlaceholder
}

type patternMatcherOption byte

const (
	patternMatcherOptionAny    = patternMatcherOption(0)
	patternMatcherOptionFull   = patternMatcherOption(1)
	patternMatcherOptionPrefix = patternMatcherOption(2)
	patternMatcherOptionSuffix = patternMatcherOption(3)
)

func (pm *patternMatcher) String() string {
	var a []string
	for i, sep := range pm.separators {
		a = append(a, sep)
		if i < len(pm.placeholders) {
			a = append(a, pm.placeholders[i].String())
		}
	}
	return strings.Join(a, "")
}

type patternMatcherPlaceholder int

// See appendPrettifyCollapsedNums()
const (
	patternMatcherPlaceholderUnknown  = patternMatcherPlaceholder(0)
	patternMatcherPlaceholderNum      = patternMatcherPlaceholder(1)
	patternMatcherPlaceholderUUID     = patternMatcherPlaceholder(2)
	patternMatcherPlaceholderIP4      = patternMatcherPlaceholder(3)
	patternMatcherPlaceholderTime     = patternMatcherPlaceholder(4)
	patternMatcherPlaceholderDate     = patternMatcherPlaceholder(5)
	patternMatcherPlaceholderDateTime = patternMatcherPlaceholder(6)
	patternMatcherPlaceholderWord     = patternMatcherPlaceholder(7)
)

func getPatternMatcherPlaceholder(s string) patternMatcherPlaceholder {
	switch s {
	case "<N>":
		return patternMatcherPlaceholderNum
	case "<UUID>":
		return patternMatcherPlaceholderUUID
	case "<IP4>":
		return patternMatcherPlaceholderIP4
	case "<TIME>":
		return patternMatcherPlaceholderTime
	case "<DATE>":
		return patternMatcherPlaceholderDate
	case "<DATETIME>":
		return patternMatcherPlaceholderDateTime
	case "<W>":
		return patternMatcherPlaceholderWord
	default:
		return patternMatcherPlaceholderUnknown
	}
}

func (ph patternMatcherPlaceholder) String() string {
	switch ph {
	case patternMatcherPlaceholderUnknown:
		return "<UNKNOWN>"
	case patternMatcherPlaceholderNum:
		return "<N>"
	case patternMatcherPlaceholderUUID:
		return "<UUID>"
	case patternMatcherPlaceholderIP4:
		return "<IP4>"
	case patternMatcherPlaceholderTime:
		return "<TIME>"
	case patternMatcherPlaceholderDate:
		return "<DATE>"
	case patternMatcherPlaceholderDateTime:
		return "<DATETIME>"
	case patternMatcherPlaceholderWord:
		return "<W>"
	default:
		logger.Panicf("BUG: unexpected placeholder=%d", ph)
		return ""
	}
}

func newPatternMatcher(s string, pmo patternMatcherOption) *patternMatcher {
	var separators []string
	var placeholders []patternMatcherPlaceholder

	offset := 0
	separator := ""
	for offset < len(s) {
		n := strings.IndexByte(s[offset:], '<')
		if n < 0 {
			separator += s[offset:]
			break
		}
		separator += s[offset : offset+n]
		offset += n

		n = strings.IndexByte(s[offset:], '>')
		if n < 0 {
			separator += s[offset:]
			break
		}
		placeholder := s[offset : offset+n+1]
		offset += n + 1

		ph := getPatternMatcherPlaceholder(placeholder)
		if ph == patternMatcherPlaceholderUnknown {
			separator += placeholder
			continue
		}

		separators = append(separators, separator)
		placeholders = append(placeholders, ph)
		separator = ""
	}
	separators = append(separators, separator)

	return &patternMatcher{
		pmo:          pmo,
		separators:   separators,
		placeholders: placeholders,
	}
}
