package logstorage

import (
	"github.com/valyala/quicktemplate"
)

// Field is a single field for the log entry.
type Field struct {
	// Name is the name of the field
	Name string

	// Value is the value of the field
	Value string
}

func (f *Field) marshalToJSON(dst []byte) []byte {
	name := f.Name
	if name == "" {
		name = "_msg"
	}
	dst = quicktemplate.AppendJSONString(dst, name, true)
	dst = append(dst, ':')
	dst = quicktemplate.AppendJSONString(dst, f.Value, true)
	return dst
}

// MarshalFieldsToJSON appends JSON-marshaled fields to dst and returns the result.
func MarshalFieldsToJSON(dst []byte, fields []Field) []byte {
	fields = SkipLeadingFieldsWithoutValues(fields)
	dst = append(dst, '{')
	if len(fields) > 0 {
		dst = fields[0].marshalToJSON(dst)
		fields = fields[1:]
		for i := range fields {
			f := &fields[i]
			if f.Value == "" {
				// Skip fields without values
				continue
			}
			dst = append(dst, ',')
			dst = f.marshalToJSON(dst)
		}
	}
	dst = append(dst, '}')
	return dst
}

// SkipLeadingFieldsWithoutValues skips leading fields without values.
func SkipLeadingFieldsWithoutValues(fields []Field) []Field {
	i := 0
	for i < len(fields) && fields[i].Value == "" {
		i++
	}
	return fields[i:]
}
