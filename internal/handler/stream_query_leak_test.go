package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/proxy"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamQuery Memory Leak Tests", func() {
	var (
		testServer  *httptest.Server
		backends    []*httptest.Server
		serverGroup []servergroup.Server
	)

	closeServers := func() {
		if testServer != nil {
			testServer.Close()
		}
		for _, backend := range backends {
			if backend != nil {
				backend.Close()
			}
		}
		backends = nil
		serverGroup = nil
	}

	setupBackends := func(numBackends int, dataSize int) {
		backends = make([]*httptest.Server, numBackends)
		serverGroup = make([]servergroup.Server, numBackends)

		for i := range numBackends {
			serverNum := i
			backends[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := strings.Repeat(fmt.Sprintf(`{"_time":1234567890.123,"_msg":"test from server%d"}`+"\n", serverNum+1), dataSize)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(data))
			}))

			serverGroup[i] = servergroup.Server{
				ClusterName: fmt.Sprintf("test-server-%d", i+1),
				Target:      strings.TrimPrefix(backends[i].URL, "http://"),
				Scheme:      "http",
			}
		}
	}

	setupTestServer := func() {
		router := gin.New()
		router.GET("/select/logsql/query", func(c *gin.Context) {
			query := handler.NewStreamQuery()
			proxyInstance := proxy.NewStreamProxy[[]byte](serverGroup, http.DefaultClient, c)
			proxyInstance.ProxyRequest(query)
		})
		testServer = httptest.NewServer(router)
	}

	performRequests := func(iterations int) {
		for i := range iterations {
			resp, err := http.Get(testServer.URL + "/select/logsql/query?query=*&limit=1000")
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Iteration %d: request failed", i+1))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Iteration %d: failed to read response body", i+1))
			Expect(body).NotTo(BeEmpty())

			err = resp.Body.Close()
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Iteration %d: failed to close response body", i+1))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	}

	checkMemoryLeak := func(initialGoroutines int, initialMemStats runtime.MemStats) {
		runtime.GC()

		var finalMemStats runtime.MemStats
		runtime.ReadMemStats(&finalMemStats)
		finalGoroutines := runtime.NumGoroutine()

		heapObjectsIncrease := int64(finalMemStats.HeapObjects) - int64(initialMemStats.HeapObjects)
		goroutineIncrease := finalGoroutines - initialGoroutines

		Expect(finalGoroutines).To(BeNumerically("<=", initialGoroutines),
			fmt.Sprintf("Goroutine leak: %d goroutines leaked", goroutineIncrease))

		Expect(heapObjectsIncrease).To(BeNumerically("<=", finalMemStats.HeapObjects),
			fmt.Sprintf("Heap objects leak: increased by %d (max: %d)", heapObjectsIncrease, finalMemStats.HeapObjects))
	}

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
	})

	AfterEach(func() {
		closeServers()
	})

	Context("with single backend server", func() {
		It("should not leak memory with small data size", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(1, 1000)
			setupTestServer()
			performRequests(2)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})

		It("should not leak memory with large data size", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(1, 10000)
			setupTestServer()
			performRequests(2)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})
	})

	Context("with multiple backend servers", func() {
		It("should not leak memory with 3 backends", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(3, 5000)
			setupTestServer()
			performRequests(2)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})

		It("should not leak memory with 10 backends", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(10, 2000)
			setupTestServer()
			performRequests(2)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})
	})

	Context("with many iterations", func() {
		It("should not leak memory after 10 requests", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(3, 5000)
			setupTestServer()
			performRequests(10)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})

		It("should not leak memory after 100 requests", func() {
			var initialMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)
			initialGoroutines := runtime.NumGoroutine()

			setupBackends(3, 3000)
			setupTestServer()
			performRequests(100)

			closeServers()
			checkMemoryLeak(initialGoroutines, initialMemStats)
		})
	})
})
