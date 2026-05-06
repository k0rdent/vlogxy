package logstorage

import (
	"fmt"
	"html"
	"strings"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pattern represents text pattern in the form 'some_text<some_field>other_text...'
type pattern struct {
	// steps contains steps for extracting fields from string
	steps []patternStep

	// matches contains matches for every step in steps
	matches []string

	// fields contains matches for non-empty fields
	fields []patternField
}

type patternField struct {
	name  string
	value *string
}

type patternStep struct {
	prefix string

	field    string
	fieldOpt string
}

func parsePattern(s string) (*pattern, error) {
	steps, err := parsePatternSteps(s)
	if err != nil {
		return nil, err
	}

	// Verify that prefixes are non-empty between fields. The first prefix may be empty.
	for i := 1; i < len(steps); i++ {
		if steps[i].prefix == "" {
			return nil, fmt.Errorf("missing delimiter between <%s> and <%s>", steps[i-1].field, steps[i].field)
		}
	}

	// Verify that fields do not end with '*'
	for _, step := range steps {
		if prefixfilter.IsWildcardFilter(step.field) {
			return nil, fmt.Errorf("wildcard field %q isn't supported", step.field)
		}
	}

	// Build pattern struct

	matches := make([]string, len(steps))

	var fields []patternField
	for i, step := range steps {
		if step.field != "" {
			fields = append(fields, patternField{
				name:  step.field,
				value: &matches[i],
			})
		}
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("pattern %q must contain at least a single named field in the form <field_name>", s)
	}

	ptn := &pattern{
		steps:   steps,
		matches: matches,
		fields:  fields,
	}
	return ptn, nil
}

func parsePatternSteps(s string) ([]patternStep, error) {
	steps, err := parsePatternStepsInternal(s)
	if err != nil {
		return nil, err
	}

	// unescape prefixes
	for i := range steps {
		step := &steps[i]
		step.prefix = html.UnescapeString(step.prefix)
	}

	// extract options part from fields
	for i := range steps {
		step := &steps[i]
		field := step.field
		if n := strings.IndexByte(field, ':'); n >= 0 {
			step.fieldOpt = strings.TrimSpace(field[:n])
			field = field[n+1:]
		}
		step.field = strings.TrimSpace(field)
	}

	return steps, nil
}

func parsePatternStepsInternal(s string) ([]patternStep, error) {
	if len(s) == 0 {
		return nil, nil
	}

	var steps []patternStep

	n := strings.IndexByte(s, '<')
	if n < 0 {
		steps = append(steps, patternStep{
			prefix: s,
		})
		return steps, nil
	}
	prefix := s[:n]
	s = s[n+1:]
	for {
		n := strings.IndexByte(s, '>')
		if n < 0 {
			return nil, fmt.Errorf("missing '>' for <%s", s)
		}
		field := s[:n]
		s = s[n+1:]

		if field == "_" || field == "*" {
			field = ""
		}
		steps = append(steps, patternStep{
			prefix: prefix,
			field:  field,
		})
		if len(s) == 0 {
			break
		}

		n = strings.IndexByte(s, '<')
		if n < 0 {
			steps = append(steps, patternStep{
				prefix: s,
			})
			break
		}
		prefix = s[:n]
		s = s[n+1:]
	}

	return steps, nil
}
