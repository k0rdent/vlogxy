package handler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/pkg/common"
)

var _ = Describe("Stats Merge method tests", func() {
	Context("when merging empty responses", func() {
		It("should return empty result", func() {
			stats := handler.NewStatsQuery(countPipe())
			result, err := stats.Merge([]handler.StatsResponse{})
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Data.Result).To(BeEmpty())
		})
	})

	Context("when merging single response", func() {
		It("should return the same data", func() {
			stats := handler.NewStatsQuery(countPipe())
			responses := []handler.StatsResponse{
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "100"},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := stats.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Status).To(Equal("success"))
			Expect(response.Data.ResultType).To(Equal("vector"))
			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Value[1]).To(Equal("100"))
		})
	})

	Context("when merging multiple responses with same metrics", func() {
		It("should aggregate values for the same metric", func() {
			stats := handler.NewStatsQuery(countPipe("label"))
			responses := []handler.StatsResponse{
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"label": "value1", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "100"},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"label": "value1", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "200"},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := stats.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Metric).To(HaveKeyWithValue("label", "value1"))
			Expect(response.Data.Result[0].Value[1]).To(Equal("300"))
		})
	})

	Context("when merging multiple responses with different metrics", func() {
		It("should keep all metrics separate", func() {
			stats := handler.NewStatsQuery(countPipe("label"))
			responses := []handler.StatsResponse{
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"label": "value1", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "100"},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"label": "value2", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "200"},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := stats.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(2))
		})
	})

	Context("when merging responses with complex metrics", func() {
		It("should handle multiple labels correctly", func() {
			stats := handler.NewStatsQuery(countPipe("app", "env"))
			responses := []handler.StatsResponse{
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"app": "web", "env": "prod", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "50"},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                `json:"resultType"`
						Result     []handler.StatsSeries `json:"result"`
					}{
						ResultType: "vector",
						Result: []handler.StatsSeries{
							{
								Metric: map[string]string{"app": "web", "env": "prod", "__name__": "value"},
								Value:  common.ValuePair{1234567890.0, "150"},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := stats.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Metric).To(HaveKeyWithValue("app", "web"))
			Expect(response.Data.Result[0].Metric).To(HaveKeyWithValue("env", "prod"))
			Expect(response.Data.Result[0].Value[1]).To(Equal("200"))
		})
	})
})
