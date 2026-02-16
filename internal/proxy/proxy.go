package proxy

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	log "github.com/sirupsen/logrus"
)

type ProxyGroup[T any] interface {
	ProxyRequest(interfaces.ResponseAggregator[T])
}

type Proxy[T any] struct {
	serverGroup []servergroup.Server
	httpClient  interfaces.HTTPClient
	ginContext  *gin.Context
}

func NewProxy[T any](serverGroup []servergroup.Server, httpClient interfaces.HTTPClient, c *gin.Context) ProxyGroup[T] {
	return &Proxy[T]{
		serverGroup: serverGroup,
		httpClient:  httpClient,
		ginContext:  c,
	}
}

func (p *Proxy[T]) ProxyRequest(aggregator interfaces.ResponseAggregator[T]) {
	ctx := p.ginContext.Request.Context()
	respChan := collectResponses(ctx, p.serverGroup, p.ginContext.Request.URL)
	result := make([]T, 0, len(p.serverGroup))
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for resp := range respChan {
		wg.Go(func() {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Errorf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				log.Errorf("received non-200 status code %d from backend", resp.StatusCode)
				return
			}

			parsedResp, err := aggregator.ParseResponse(resp)
			if err != nil {
				log.Errorf("failed to parse response: %v", err)
				return
			}

			mu.Lock()
			result = append(result, parsedResp)
			mu.Unlock()
		})
	}
	wg.Wait()

	output, err := aggregator.Merge(result)
	if err != nil {
		log.Errorf("failed to merge responses: %v", err)
		p.ginContext.JSON(http.StatusInternalServerError, gin.H{"error": "failed to merge responses"})
		return
	}

	p.ginContext.Data(http.StatusOK, "application/json", output)
}
