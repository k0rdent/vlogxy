package logstorage

import (
	"strings"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// filterStreamID is the filter for `_stream_id:id`
type filterStreamID struct {
	streamIDs []streamID

	// If q is non-nil, then streamIDs must be populated from q before filter execution.
	q *Query

	// qFieldName must be set to field name for obtaining values from if q is non-nil.
	qFieldName string
}

func newFilterStreamID(streamIDs []streamID) *filterStreamID {
	return &filterStreamID{
		streamIDs: streamIDs,
	}
}

func newFilterStreamIDFromQuery(q *Query, qFieldName string) *filterStreamID {
	return &filterStreamID{
		q:          q,
		qFieldName: qFieldName,
	}
}

func (fs *filterStreamID) String() string {
	if fs.q != nil {
		return "_stream_id:in(" + fs.q.String() + ")"
	}

	streamIDs := fs.streamIDs
	if len(streamIDs) == 1 {
		return "_stream_id:" + string(streamIDs[0].marshalString(nil))
	}

	a := make([]string, len(streamIDs))
	for i, streamID := range streamIDs {
		a[i] = string(streamID.marshalString(nil))
	}
	return "_stream_id:in(" + strings.Join(a, ",") + ")"
}

func (fs *filterStreamID) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_stream_id")
}
