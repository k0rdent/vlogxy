package handler

import (
	"bufio"
	"encoding/json"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

type Logs []Log
type Log map[string]any

type Query struct {
	*common.RequestPath
}

func NewQuery() interfaces.ResponseAggregator[Logs] {
	return &Query{}
}

func (q *Query) ParseResponse(resp *http.Response) (Logs, error) {
	scanner := bufio.NewScanner(resp.Body)

	logs := Logs{}
	for scanner.Scan() {
		line := scanner.Bytes()

		var m Log
		if err := json.Unmarshal(line, &m); err != nil {
			log.Errorf("failed to unmarshal log line: %v", err)
			continue
		}

		logs = append(logs, m)
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("error reading response: %v", err)
		return nil, err
	}

	return logs, nil
}

func (q *Query) Merge(responses []Logs) ([]byte, error) {
	var rawOutput []byte

	for _, logs := range responses {
		for _, vlLog := range logs {
			buf, err := marshalQuery(vlLog)
			if err != nil {
				log.Errorf("failed to marshal log: %v", err)
				continue
			}
			rawOutput = append(rawOutput, buf...)
		}
	}
	return rawOutput, nil
}

func marshalQuery(m Log) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}
