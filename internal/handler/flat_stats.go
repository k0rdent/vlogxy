package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/internal/merger"
	"github.com/k0rdent/vlogxy/internal/parser"
	log "github.com/sirupsen/logrus"
)

// FlatResponse holds a collection of flat JSON objects returned by a stats query
// routed through the /query endpoint. Each line in the backend response is one object.
type FlatResponse []map[string]string

// FlatStatsQuery handles aggregation when a LogsQL query containing stats functions
// is sent to the /query endpoint.
type FlatStatsQuery struct {
	pipes []*parser.Pipe
}

// NewFlatStatsQuery creates a new FlatStatsQuery aggregator for the given query.
func NewFlatStatsQuery(pipes []*parser.Pipe) interfaces.ResponseAggregator[FlatResponse] {
	return &FlatStatsQuery{pipes: pipes}
}

func (f *FlatStatsQuery) ParseResponse(resp *http.Response) (FlatResponse, error) {
	var result FlatResponse

	decoder := json.NewDecoder(resp.Body)
	for {
		var obj map[string]string
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("flat stats: failed to decode line: %v", err)
			continue
		}
		result = append(result, obj)
	}
	return result, nil
}

func (f *FlatStatsQuery) Merge(responses []FlatResponse) ([]byte, error) {
	totalLen := 0
	for _, resp := range responses {
		totalLen += len(resp)
	}
	rows := make([]map[string]string, 0, totalLen)

	// Flatten all backend responses into a single row slice.
	for _, resp := range responses {
		rows = append(rows, resp...)
	}

	// Apply each pipe's merge logic, sorted by registered Order.
	var err error
	for _, task := range merger.OrderedPipeTasks(f.pipes) {
		rows, err = task.Merge(task.Pipe, rows)
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	for _, row := range rows {
		line, err := json.Marshal(row)
		if err != nil {
			log.Errorf("flat stats: failed to marshal row: %v", err)
			continue
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}
