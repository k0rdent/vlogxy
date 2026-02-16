package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/proxy"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
)

func getFakeServers(count uint) []servergroup.Server {
	servers := make([]servergroup.Server, count)
	for i := range servers {
		servers[i] = servergroup.Server{
			ClusterName: fmt.Sprintf("server%d", i+1),
			Target:      fmt.Sprintf("localhost:808%d", i+1),
			Scheme:      "http",
		}
	}

	return servers
}

func TestStreamQueryMemoryLeak(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	// Create multiple backend servers
	numBackends := 3
	vlBackends := make([]*httptest.Server, numBackends)
	servers := make([]servergroup.Server, numBackends)

	for i := range vlBackends {
		serverNum := i
		vlBackends[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := strings.Repeat(fmt.Sprintf(`{"_time":1234567890.123,"_msg":"test from server%d"}`+"\n", serverNum+1), 10000)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(data))
		}))

		servers[i] = servergroup.Server{
			ClusterName: fmt.Sprintf("test-server-%d", i+1),
			Target:      strings.TrimPrefix(vlBackends[i].URL, "http://"),
			Scheme:      "http",
		}
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/select/logsql/query", func(c *gin.Context) {
		query := NewStreamQuery()
		proxyInstance := proxy.NewStreamProxy[[]byte](servers, http.DefaultClient, c)
		proxyInstance.ProxyRequest(query)
	})
	testServer := httptest.NewServer(router)

	iterations := 10
	for i := range iterations {
		resp, err := http.Get(testServer.URL + "/select/logsql/query?query=*&limit=1000")
		if err != nil {
			t.Errorf("Iteration %d: request failed: %v", i+1, err)
			continue
		}

		if _, err = io.ReadAll(resp.Body); err != nil {
			t.Errorf("Iteration %d: failed to read response body: %v", i+1, err)
		}

		if err := resp.Body.Close(); err != nil {
			t.Errorf("Iteration %d: failed to close response body: %v", i+1, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Logf("Iteration %d: unexpected status code: %d", i+1, resp.StatusCode)
		}
	}

	testServer.Close()
	for _, vlBackend := range vlBackends {
		vlBackend.Close()
	}

	runtime.GC()
	runtime.Gosched()

	finalGoroutines := runtime.NumGoroutine()

	if finalGoroutines > initialGoroutines {
		t.Errorf("Potential goroutine leak: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	} else {
		t.Logf("No leak: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	}
}
