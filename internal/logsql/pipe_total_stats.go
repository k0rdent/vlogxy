package logsql

import (
	"fmt"
)

func parsePipeTotalStats(lex *lexer) (pipe, error) {
	if !lex.isKeyword(totalStats) {
		return nil, fmt.Errorf("expecting `%s`; got %q", totalStats, lex.token)
	}
	lex.nextToken()

	return parsePipeRunningStatsExt(lex, totalStats)
}
