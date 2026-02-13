package handler

import (
	"bufio"
	"context"
	"net/http"

	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/pkg/common"
	log "github.com/sirupsen/logrus"
)

type StreamQuery struct {
	*common.RequestPath
}

func NewStreamQuery(path, rawQuery string) interfaces.StreamResponseAggregator[[]byte] {
	return &StreamQuery{
		RequestPath: &common.RequestPath{
			Path:     path,
			RawQuery: rawQuery,
		},
	}
}

func (s *StreamQuery) StreamParseResponse(ctx context.Context, resp *http.Response) (<-chan []byte, error) {
	dataChan := make(chan []byte)

	go func() {
		defer close(dataChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		defer func() {
			if err := scanner.Err(); err != nil {
				log.Errorf("error reading response: %v", err)
			}
		}()

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				data := make([]byte, len(scanner.Bytes()))
				copy(data, scanner.Bytes())
				dataChan <- data
			}
		}
	}()

	return dataChan, nil
}

func (s *StreamQuery) GetURL(scheme string, host string, path string) string {
	return common.BuildURL(scheme, host, path+s.Path, s.RawQuery)
}
