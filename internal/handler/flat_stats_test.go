package handler_test

import (
	"bufio"
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/parser"
)

// parseNDJSON splits newline-delimited JSON output into a slice of string maps.
func parseNDJSON(data []byte) ([]map[string]string, error) {
	var result []map[string]string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var obj map[string]string
		if err := json.Unmarshal(line, &obj); err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, scanner.Err()
}

var _ = Describe("FlatStatsQuery Merge method tests", func() {
	Describe("empty input", func() {
		It("returns empty output when no responses provided", func() {
			pipes := []*parser.Pipe{
				{Name: "stats", ByFields: []string{"service"}, Funcs: []*parser.Func{{Name: "count", ResultName: "total"}}},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			out, err := agg.Merge([]handler.FlatResponse{})
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(BeEmpty())
		})

		It("returns empty output when all responses are empty", func() {
			pipes := []*parser.Pipe{
				{Name: "stats", ByFields: []string{"service"}, Funcs: []*parser.Func{{Name: "count", ResultName: "total"}}},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			out, err := agg.Merge([]handler.FlatResponse{{}, {}})
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(BeEmpty())
		})
	})

	Describe("single response", func() {
		It("returns the same record unchanged", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "total": "42"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["service"]).To(Equal("api"))
			Expect(rows[0]["total"]).To(Equal("42"))
		})
	})

	Describe("sum aggregation (count)", func() {
		It("sums count values for the same group key across backends", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "total": "100"}},
				{{"service": "api", "total": "200"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["service"]).To(Equal("api"))
			Expect(rows[0]["total"]).To(Equal("300"))
		})

		It("sums sum values for the same group key across backends", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"env"},
					Funcs:    []*parser.Func{{Name: "sum", ResultName: "bytes"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"env": "prod", "bytes": "1000"}},
				{{"env": "prod", "bytes": "2000"}},
				{{"env": "prod", "bytes": "500"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["bytes"]).To(Equal("3500"))
		})
	})

	Describe("min aggregation", func() {
		It("takes the minimum value across backends", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "min", ResultName: "latency_min"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "latency_min": "50"}},
				{{"service": "api", "latency_min": "30"}},
				{{"service": "api", "latency_min": "70"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["latency_min"]).To(Equal("30"))
		})
	})

	Describe("max aggregation", func() {
		It("takes the maximum value across backends", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "max", ResultName: "latency_max"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "latency_max": "50"}},
				{{"service": "api", "latency_max": "30"}},
				{{"service": "api", "latency_max": "70"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["latency_max"]).To(Equal("70"))
		})
	})

	Describe("avg aggregation", func() {
		It("averages values across backends", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "avg", ResultName: "latency_avg"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "latency_avg": "100"}},
				{{"service": "api", "latency_avg": "200"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["latency_avg"]).To(Equal("150"))
		})
	})

	Describe("different group keys", func() {
		It("keeps records with different by-field values separate", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{
					{"service": "api", "total": "100"},
					{"service": "worker", "total": "50"},
				},
				{
					{"service": "api", "total": "200"},
					{"service": "worker", "total": "75"},
				},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(2))

			byService := make(map[string]string)
			for _, row := range rows {
				byService[row["service"]] = row["total"]
			}
			Expect(byService["api"]).To(Equal("300"))
			Expect(byService["worker"]).To(Equal("125"))
		})
	})

	Describe("composite group key (multiple by fields)", func() {
		It("groups by all by-fields together", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service", "env"},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{
					{"service": "api", "env": "prod", "total": "100"},
					{"service": "api", "env": "staging", "total": "10"},
				},
				{
					{"service": "api", "env": "prod", "total": "150"},
					{"service": "api", "env": "staging", "total": "20"},
				},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(2))

			type key struct{ service, env string }
			byKey := make(map[key]string)
			for _, row := range rows {
				byKey[key{row["service"], row["env"]}] = row["total"]
			}
			Expect(byKey[key{"api", "prod"}]).To(Equal("250"))
			Expect(byKey[key{"api", "staging"}]).To(Equal("30"))
		})
	})

	Describe("multiple stats functions", func() {
		It("aggregates each function independently", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs: []*parser.Func{
						{Name: "count", ResultName: "total"},
						{Name: "min", ResultName: "latency_min"},
						{Name: "max", ResultName: "latency_max"},
					},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"service": "api", "total": "10", "latency_min": "50", "latency_max": "200"}},
				{{"service": "api", "total": "20", "latency_min": "30", "latency_max": "250"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["total"]).To(Equal("30"))
			Expect(rows[0]["latency_min"]).To(Equal("30"))
			Expect(rows[0]["latency_max"]).To(Equal("250"))
		})
	})

	Describe("no by fields (global aggregation)", func() {
		It("aggregates all records into a single group", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			responses := []handler.FlatResponse{
				{{"total": "100"}},
				{{"total": "200"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows).To(HaveLen(1))
			Expect(rows[0]["total"]).To(Equal("300"))
		})
	})

	Describe("records missing some fields", func() {
		It("skips missing by-field entries and still processes present ones", func() {
			pipes := []*parser.Pipe{
				{
					Name:     "stats",
					ByFields: []string{"service"},
					Funcs:    []*parser.Func{{Name: "count", ResultName: "total"}},
				},
			}
			agg := handler.NewFlatStatsQuery(pipes)

			// Second response is missing "service" — its record still lands in the
			// empty-key group (no by-field value), separate from "api".
			responses := []handler.FlatResponse{
				{{"service": "api", "total": "100"}},
				{{"total": "50"}},
			}

			out, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())
			rows, err := parseNDJSON(out)
			Expect(err).NotTo(HaveOccurred())
			// Two distinct groups: {service:"api"} and {} (no service).
			Expect(rows).To(HaveLen(2))
		})
	})
})
