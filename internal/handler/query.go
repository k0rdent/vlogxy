package handler

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/internal/victorialogs"
	log "github.com/sirupsen/logrus"
)

type Logs []Log
type Log map[string]any

type Query struct {
	Path     string
	RawQuery string
}

func NewQuery(rawPath, rawQuery string) victorialogs.Querier[Logs] {
	return &Query{
		Path:     rawPath,
		RawQuery: rawQuery,
	}
}

func ProxyQuery(c *gin.Context) {
	query := NewQuery(c.Request.URL.Path, c.Request.URL.RawQuery)
	response, err := victorialogs.MultiClusterProxy(query)
	if err != nil {
		log.Errorf("failed to make query: %v", err)
		c.JSON(500, gin.H{
			"message": "Failed to process proxy query",
		})
		return
	}
	c.Data(200, "application/json", response)
}

func (q *Query) ResponseHandler(resp *http.Response) (Logs, error) {
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
	}

	return logs, nil
}

func (q *Query) Merge(mergedLogs []Logs) ([]byte, error) {
	var rawOutput []byte
	for _, logs := range mergedLogs {
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

func (q *Query) GetEndpoint(scheme, target string) string {
	url := url.URL{
		Scheme:   scheme,
		Host:     target,
		Path:     q.Path,
		RawQuery: q.RawQuery,
	}
	return url.String()
}

func marshalQuery(m Log) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}
