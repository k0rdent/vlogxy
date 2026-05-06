package logstorage

import (
	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterStream is the filter for `{}` aka `_stream:{...}`
type filterStream struct {
	// f is the filter to apply
	f *StreamFilter
}

func newFilterStream(f *StreamFilter) *filterStream {
	return &filterStream{
		f: f,
	}
}

func (fs *filterStream) String() string {
	return fs.f.String()
}

func (fs *filterStream) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_stream")
}
