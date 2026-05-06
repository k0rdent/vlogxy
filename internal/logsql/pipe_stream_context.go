package logstorage

import (
	"fmt"
	"math"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

// pipeStreamContextDefaultTimeWindow is the default time window to search for surrounding logs in `stream_context` pipe.
const pipeStreamContextDefaultTimeWindow = int64(nsecsPerHour)

// pipeStreamContext processes '| stream_context ...' queries.
//
// See https://docs.victoriametrics.com/victorialogs/logsql/#stream_context-pipe
type pipeStreamContext struct {
	// linesBefore is the number of lines to return before the matching line
	linesBefore int

	// linesAfter is the number of lines to return after the matching line
	linesAfter int

	// timeWindow is the time window in nanoseconds for searching for surrounding logs
	timeWindow int64
}

func (pc *pipeStreamContext) String() string {
	s := "stream_context"
	if pc.linesBefore > 0 {
		s += fmt.Sprintf(" before %d", pc.linesBefore)
	}
	if pc.linesAfter > 0 {
		s += fmt.Sprintf(" after %d", pc.linesAfter)
	}
	if pc.linesBefore <= 0 && pc.linesAfter <= 0 {
		s += " after 0"
	}
	if pc.timeWindow != pipeStreamContextDefaultTimeWindow {
		s += " time_window " + string(marshalDurationString(nil, pc.timeWindow))
	}
	return s
}

func (pc *pipeStreamContext) Name() string {
	return "stream_context"
}

func (pc *pipeStreamContext) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter("_time")
	pf.AddAllowFilter("_stream_id")
}

func (pc *pipeStreamContext) visitSubqueries(_ func(q *Query)) {
	// nothing to do
}

func parsePipeStreamContext(lex *lexer) (pipe, error) {
	if !lex.isKeyword("stream_context") {
		return nil, fmt.Errorf("expecting 'stream_context'; got %q", lex.token)
	}
	lex.nextToken()

	linesBefore, linesAfter, err := parsePipeStreamContextBeforeAfter(lex)
	if err != nil {
		return nil, err
	}

	timeWindow := pipeStreamContextDefaultTimeWindow
	if lex.isKeyword("time_window") {
		lex.nextToken()

		token, err := lex.nextCompoundToken()
		if err != nil {
			return nil, fmt.Errorf("cannot parse 'time_window': %w", err)
		}

		d, ok := tryParseDuration(token)
		if !ok {
			return nil, fmt.Errorf("cannot parse 'time_window %s'; it must contain valid duration", token)
		}
		if timeWindow <= 0 {
			return nil, fmt.Errorf("'time_window' must be positive; got %s", token)
		}
		timeWindow = d
	}

	pc := &pipeStreamContext{
		linesBefore: linesBefore,
		linesAfter:  linesAfter,
		timeWindow:  timeWindow,
	}
	return pc, nil
}

func parsePipeStreamContextBeforeAfter(lex *lexer) (int, int, error) {
	linesBefore := 0
	linesAfter := 0
	beforeSet := false
	afterSet := false
	for {
		switch {
		case lex.isKeyword("before"):
			lex.nextToken()
			f, s, err := parseNumber(lex)
			if err != nil {
				return 0, 0, fmt.Errorf("cannot parse 'before' value in 'stream_context': %w", err)
			}
			if f < 0 {
				return 0, 0, fmt.Errorf("'before' value cannot be smaller than 0; got %q", s)
			}
			if f > math.MaxInt {
				return 0, 0, fmt.Errorf("'before' value cannot be bigger than %v; got %v", math.MaxInt, f)
			}
			linesBefore = int(f)
			beforeSet = true
		case lex.isKeyword("after"):
			lex.nextToken()
			f, s, err := parseNumber(lex)
			if err != nil {
				return 0, 0, fmt.Errorf("cannot parse 'after' value in 'stream_context': %w", err)
			}
			if f < 0 {
				return 0, 0, fmt.Errorf("'after' value cannot be smaller than 0; got %q", s)
			}
			if f > math.MaxInt {
				return 0, 0, fmt.Errorf("'after' value cannot be bigger than %v; got %v", math.MaxInt, f)
			}
			linesAfter = int(f)
			afterSet = true
		default:
			if !beforeSet && !afterSet {
				return 0, 0, fmt.Errorf("missing 'before N' or 'after N' in 'stream_context'")
			}
			return linesBefore, linesAfter, nil
		}
	}
}
