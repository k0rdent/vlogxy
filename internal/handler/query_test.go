package handler_test

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
)

var _ = Describe("Query Merge method tests", func() {
	var query *handler.Query

	BeforeEach(func() {
		query = handler.NewQuery("/select/logsql/query", "query=test").(*handler.Query)
	})

	Context("when merging empty responses", func() {
		It("should return empty output", func() {
			result, err := query.Merge([]handler.Logs{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	Context("when merging single response", func() {
		It("should return logs with newline separators", func() {
			logs := handler.Logs{
				handler.Log{"timestamp": "2024-01-01T00:00:00Z", "message": "test log 1"},
				handler.Log{"timestamp": "2024-01-01T00:00:01Z", "message": "test log 2"},
			}

			result, err := query.Merge([]handler.Logs{logs})
			Expect(err).NotTo(HaveOccurred())

			lines := strings.Split(strings.TrimSpace(string(result)), "\n")
			Expect(lines).To(HaveLen(2))

			var log1 handler.Log
			err = json.Unmarshal([]byte(lines[0]), &log1)
			Expect(err).NotTo(HaveOccurred())
			Expect(log1).To(HaveKeyWithValue("message", "test log 1"))

			var log2 handler.Log
			err = json.Unmarshal([]byte(lines[1]), &log2)
			Expect(err).NotTo(HaveOccurred())
			Expect(log2).To(HaveKeyWithValue("message", "test log 2"))
		})
	})

	Context("when merging multiple responses", func() {
		It("should concatenate all logs", func() {
			logs1 := handler.Logs{
				handler.Log{"timestamp": "2024-01-01T00:00:00Z", "message": "log from server 1"},
			}
			logs2 := handler.Logs{
				handler.Log{"timestamp": "2024-01-01T00:00:01Z", "message": "log from server 2"},
			}
			logs3 := handler.Logs{
				handler.Log{"timestamp": "2024-01-01T00:00:02Z", "message": "log from server 3"},
			}

			result, err := query.Merge([]handler.Logs{logs1, logs2, logs3})
			Expect(err).NotTo(HaveOccurred())

			lines := strings.Split(strings.TrimSpace(string(result)), "\n")
			Expect(lines).To(HaveLen(3))
		})
	})

	Context("when merging empty log collections", func() {
		It("should handle empty logs gracefully", func() {
			logs1 := handler.Logs{}
			logs2 := handler.Logs{
				handler.Log{"message": "actual log"},
			}
			logs3 := handler.Logs{}

			result, err := query.Merge([]handler.Logs{logs1, logs2, logs3})
			Expect(err).NotTo(HaveOccurred())

			lines := strings.Split(strings.TrimSpace(string(result)), "\n")
			Expect(lines).To(HaveLen(1))
		})
	})
})
