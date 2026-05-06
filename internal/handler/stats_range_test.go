package handler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/parser"
	"github.com/k0rdent/vlogxy/pkg/common"
)

// countPipe returns a stats pipe grouping by the given fields with a single count function.
func countPipe(byFields ...string) []*parser.Pipe {
	return []*parser.Pipe{
		{
			Name:     "stats",
			ByFields: byFields,
			Funcs:    []*parser.Func{{Name: "count", ResultName: "value"}},
		},
	}
}

var _ = Describe("StatsRange Merge method tests", func() {
	var statsRange *handler.StatsRange

	BeforeEach(func() {
		statsRange = handler.NewStatsRangeQuery(countPipe()).(*handler.StatsRange)
	})

	Context("when merging empty responses", func() {
		It("should return empty result", func() {
			result, err := statsRange.Merge([]handler.StatsRangeResponse{})
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsRangeResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Data.Result).To(BeEmpty())
		})
	})

	Context("when merging single response with multiple values", func() {
		It("should return values sorted by timestamp", func() {
			responses := []handler.StatsRangeResponse{
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567890.0, "100"},
									{1234567900.0, "200"},
									{1234567910.0, "300"},
								},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := statsRange.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsRangeResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Status).To(Equal("success"))
			Expect(response.Data.ResultType).To(Equal("matrix"))
			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Values).To(HaveLen(3))
		})
	})

	Context("when merging multiple responses with same metrics and timestamps", func() {
		It("should aggregate values for matching timestamps", func() {
			responses := []handler.StatsRangeResponse{
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567890.0, "100"},
									{1234567900.0, "200"},
								},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567890.0, "50"},
									{1234567900.0, "100"},
								},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := statsRange.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsRangeResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Values).To(HaveLen(2))

			for _, valuePair := range response.Data.Result[0].Values {
				ts := valuePair[0].(float64)
				val := valuePair[1].(string)

				switch ts {
				case 1234567890.0:
					Expect(val).To(Equal("150")) // 100 + 50
				case 1234567900.0:
					Expect(val).To(Equal("300")) // 200 + 100
				}
			}
		})
	})

	Context("when merging responses with different metrics", func() {
		It("should keep metrics separate", func() {
			// Pipe groups by "label", matching the metric labels in the responses.
			agg := handler.NewStatsRangeQuery(countPipe("label")).(*handler.StatsRange)
			responses := []handler.StatsRangeResponse{
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567890.0, "100"},
								},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value2"},
								Values: []common.ValuePair{
									{1234567890.0, "200"},
								},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := agg.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsRangeResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(2))
		})
	})

	Context("when merging responses with non-overlapping timestamps", func() {
		It("should include all timestamps sorted", func() {
			responses := []handler.StatsRangeResponse{
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567890.0, "100"},
									{1234567910.0, "300"},
								},
							},
						},
					},
					Status: "success",
				},
				{
					Data: struct {
						ResultType string                     `json:"resultType"`
						Result     []handler.StatsRangeSeries `json:"result"`
					}{
						ResultType: "matrix",
						Result: []handler.StatsRangeSeries{
							{
								Metric: map[string]string{"label": "value1"},
								Values: []common.ValuePair{
									{1234567900.0, "200"},
									{1234567920.0, "400"},
								},
							},
						},
					},
					Status: "success",
				},
			}

			result, err := statsRange.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.StatsRangeResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Data.Result).To(HaveLen(1))
			Expect(response.Data.Result[0].Values).To(HaveLen(4))

			// Check that values are sorted by timestamp
			timestamps := make([]float64, 0, len(response.Data.Result[0].Values))
			for _, valuePair := range response.Data.Result[0].Values {
				timestamps = append(timestamps, valuePair[0].(float64))
			}
			Expect(timestamps).To(Equal([]float64{1234567890.0, 1234567900.0, 1234567910.0, 1234567920.0}))
		})
	})
})
