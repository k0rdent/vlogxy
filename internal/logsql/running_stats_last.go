package logstorage

import (
	"fmt"
	"strconv"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type runningStatsLast struct {
	fieldName string
	offset    int
}

func (sl *runningStatsLast) String() string {
	s := "last(" + quoteTokenIfNeeded(sl.fieldName) + ")"
	if sl.offset > 0 {
		s += fmt.Sprintf(" offset %d", sl.offset)
	}
	return s
}

func (sl *runningStatsLast) updateNeededFields(pf *prefixfilter.Filter) {
	pf.AddAllowFilter(sl.fieldName)
}

func parseRunningStatsLast(lex *lexer) (runningStatsFunc, error) {
	args, err := parseStatsFuncArgs(lex, "last")
	if err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("unexpeccted number of args for the last() function; got %d; want 1; args: %q", len(args), args)
	}

	fieldName := args[0]

	offset := 0
	if lex.isKeyword("offset") {
		lex.nextToken()
		offsetStr := lex.token
		lex.nextToken()
		n, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("cannot parse offset=%q at last(%q): %w", offsetStr, fieldName, err)
		}
		if n < 0 {
			return nil, fmt.Errorf("offset=%d cannot be negative at last(%q)", n, fieldName)
		}
		offset = n
	}

	sf := &runningStatsLast{
		fieldName: fieldName,
		offset:    offset,
	}
	return sf, nil
}
