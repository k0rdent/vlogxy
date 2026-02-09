package victorialogs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/k0rdent/victorialogs-aggregator/interanl/config"
	log "github.com/sirupsen/logrus"
)

type ProxyRequest struct {
	Path       string
	RawQuery   string
	ConfigData config.ConfigData
}

type QueryResponse []string

type Querier interface {
	Query() ([]byte, error)
	Stream() (<-chan []byte, error)
}

func NewProxyRequest(rawPath, rawQuery string) Querier {
	return &ProxyRequest{
		Path:       rawPath,
		RawQuery:   rawQuery,
		ConfigData: config.GlobalConfig.GetData(),
	}
}

func (pr *ProxyRequest) Query() ([]byte, error) {
	configData := config.GlobalConfig.GetData()
	wg := sync.WaitGroup{}
	reader := make(chan QueryResponse)
	mergedLogs := QueryResponse{}

	go func() {
		for resp := range reader {
			mergedLogs = append(mergedLogs, resp...)
		}
	}()

	for _, serverGroup := range configData.ServerGroups {
		wg.Go(func() {
			url := url.URL{
				Host:     serverGroup.Target,
				Scheme:   serverGroup.Scheme,
				Path:     pr.Path,
				RawQuery: pr.RawQuery,
			}
			endpoint := url.String()

			respBody, err := forwardRequest(endpoint)
			if err != nil {
				log.Errorf("failed to forward request to %s: %v", endpoint, err)
				return
			}

			log.Infof("response body: %s", string(respBody))

			response, err := parseResponse(respBody)
			if err != nil {
				log.Errorf("failed to parse response from %s: %v", endpoint, err)
				return
			}

			log.Infof("hello world: %v", response)

			reader <- response
		})
	}

	wg.Wait()
	close(reader)

	return json.Marshal(mergedLogs)
}

func (pr *ProxyRequest) Stream() (<-chan []byte, error) {
	return nil, nil
}

func forwardRequest(endpoint string) ([]byte, error) {
	proxyReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Errorf("failed to create request for %s: %v", endpoint, err)
		return nil, err
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		log.Errorf("failed to forward request to %s: %v", endpoint, err)
		return nil, err
	}

	defer func() {
		if err := proxyResp.Body.Close(); err != nil {
			log.Errorf("failed to close response body from %s: %v", endpoint, err)
		}
	}()

	respBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		log.Errorf("failed to read response body from %s: %v", endpoint, err)
	}

	return respBody, nil
}

func parseResponse(data []byte) ([]string, error) {
	var msgs []string

	lines := bytes.SplitSeq(data, []byte{'\n'})

	for line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] != '{' {
			continue
		}

		var m struct {
			Msg string `json:"_msg"`
		}

		if err := json.Unmarshal(line, &m); err == nil && m.Msg != "" {
			msgs = append(msgs, m.Msg)
		}
	}

	return msgs, nil
}
