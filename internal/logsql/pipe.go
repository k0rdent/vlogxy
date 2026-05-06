package logstorage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/VictoriaLogs/lib/prefixfilter"
)

type pipe interface {
	// Name returns the name of the pipe.
	Name() string

	// String returns string representation of the pipe.
	String() string

	// updateNeededFields must update pf with fields it needs and not needs at the input.
	updateNeededFields(pf *prefixfilter.Filter)

	// visitSubqueries must call visitFunc for all the subqueries, which exist at the pipe (recursively).
	visitSubqueries(visitFunc func(q *Query))
}

func parsePipes(lex *lexer) ([]pipe, error) {
	var pipes []pipe
	for {
		p, err := parsePipe(lex)
		if err != nil {
			return nil, err
		}
		pipes = append(pipes, p)

		switch {
		case lex.isKeyword("|"):
			lex.nextToken()
		case lex.isKeyword(")", ""):
			return pipes, nil
		default:
			return nil, fmt.Errorf("unexpected token after [%s]: %q; expecting '|' or ')'", pipes[len(pipes)-1], lex.token)
		}
	}
}

func parsePipe(lex *lexer) (pipe, error) {
	pps := getPipeParsers()
	for pipeName, parseFunc := range pps {
		if !lex.isKeyword(pipeName) {
			continue
		}
		p, err := parseFunc(lex)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q pipe: %w", pipeName, err)
		}
		return p, nil
	}

	lexState := lex.backupState()

	// Try parsing stats pipe without 'stats' keyword
	ps, err := parsePipeStatsNoStatsKeyword(lex)
	if err == nil {
		return ps, nil
	}
	lex.restoreState(lexState)

	// Try parsing filter pipe without 'filter' keyword
	pf, err := parsePipeFilterNoFilterKeyword(lex)
	if err == nil {
		return pf, nil
	}
	lex.restoreState(lexState)

	return nil, fmt.Errorf("unexpected pipe %q", lex.token)
}

var pipeParsers map[string]pipeParseFunc
var pipeParsersOnce sync.Once

type pipeParseFunc func(lex *lexer) (pipe, error)

func getPipeParsers() map[string]pipeParseFunc {
	pipeParsersOnce.Do(initPipeParsers)
	return pipeParsers
}

func initPipeParsers() {
	pipeParsers = map[string]pipeParseFunc{
		"block_stats":       parsePipeBlockStats,
		"blocks_count":      parsePipeBlocksCount,
		"collapse_nums":     parsePipeCollapseNums,
		"decolorize":        parsePipeDecolorize,
		"del":               parsePipeDelete,
		"delete":            parsePipeDelete,
		"drop":              parsePipeDelete,
		"drop_empty_fields": parsePipeDropEmptyFields,
		"extract":           parsePipeExtract,
		"extract_regexp":    parsePipeExtractRegexp,
		"eval":              parsePipeMath,
		"facets":            parsePipeFacets,
		"field_names":       parsePipeFieldNames,
		"field_values":      parsePipeFieldValues,
		"fields":            parsePipeFields,
		"filter":            parsePipeFilter,
		"first":             parsePipeFirst,
		"format":            parsePipeFormat,
		"generate_sequence": parsePipeGenerateSequence,
		"hash":              parsePipeHash,
		"join":              parsePipeJoin,
		"json_array_len":    parsePipeJSONArrayLen,
		"keep":              parsePipeFields,
		"last":              parsePipeLast,
		"len":               parsePipeLen,
		"math":              parsePipeMath,
		"mv":                parsePipeRename,
		"offset":            parsePipeOffset,
		"order":             parsePipeSort,
		"pack_json":         parsePipePackJSON,
		"pack_logfmt":       parsePipePackLogfmt,
		"query_stats":       parsePipeQueryStats,
		"rename":            parsePipeRename,
		"replace":           parsePipeReplace,
		"replace_regexp":    parsePipeReplaceRegexp,
		"rm":                parsePipeDelete,
		"running_stats":     parsePipeRunningStats,
		"sample":            parsePipeSample,
		"set_stream_fields": parsePipeSetStreamFields,
		"skip":              parsePipeOffset,
		"sort":              parsePipeSort,
		"split":             parsePipeSplit,
		"stats":             parsePipeStats,
		"stats_remote":      parsePipeStats,
		"stream_context":    parsePipeStreamContext,
		"time_add":          parsePipeTimeAdd,
		"top":               parsePipeTop,
		"total_stats":       parsePipeTotalStats,
		"union":             parsePipeUnion,
		"uniq":              parsePipeUniq,
		"unpack_json":       parsePipeUnpackJSON,
		"unpack_logfmt":     parsePipeUnpackLogfmt,
		"unpack_syslog":     parsePipeUnpackSyslog,
		"unpack_words":      parsePipeUnpackWords,
		"unroll":            parsePipeUnroll,
		"where":             parsePipeFilter,
	}
}

func isPipeName(s string) bool {
	pps := getPipeParsers()
	sLower := strings.ToLower(s)
	return pps[sLower] != nil
}
