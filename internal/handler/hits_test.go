package handler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
)

var _ = Describe("HitsQuery Merge method tests", func() {
	var hitsQuery *handler.HitsQuery

	BeforeEach(func() {
		hitsQuery = handler.NewHits().(*handler.HitsQuery)
	})

	Context("when merging empty responses", func() {
		It("should return empty hits array", func() {
			result, err := hitsQuery.Merge([]handler.Response{})
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.HitsArr).To(BeEmpty())
		})
	})

	Context("when merging single response", func() {
		It("should return the same hits", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api"},
							Timestamps: []string{"2024-01-01T00:00:00Z", "2024-01-01T00:01:00Z"},
							Values:     []int{10, 20},
							Total:      30,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(1))
			Expect(response.HitsArr[0].Fields).To(Equal(map[string]string{"service": "api"}))
			Expect(response.HitsArr[0].Total).To(Equal(30))
		})
	})

	Context("when merging responses with same fields", func() {
		It("should aggregate timestamps and values, and sum totals", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{10},
							Total:      10,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{20},
							Total:      20,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(1))
			Expect(response.HitsArr[0].Total).To(Equal(30)) // 10 + 20

			// Should have aggregated the values for the same timestamp
			Expect(response.HitsArr[0].Timestamps).To(HaveLen(1))
			Expect(response.HitsArr[0].Timestamps).To(ContainElement("2024-01-01T00:00:00Z"))
			Expect(response.HitsArr[0].Values).To(ContainElement(30)) // 10 + 20
		})
	})

	Context("when merging responses with different fields", func() {
		It("should keep hits separate", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{10},
							Total:      10,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "web"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{20},
							Total:      20,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(2))
		})
	})

	Context("when merging responses with multiple timestamps", func() {
		It("should aggregate values for matching timestamps", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields: map[string]string{"service": "api"},
							Timestamps: []string{
								"2024-01-01T00:00:00Z",
								"2024-01-01T00:01:00Z",
							},
							Values: []int{10, 15},
							Total:  25,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields: map[string]string{"service": "api"},
							Timestamps: []string{
								"2024-01-01T00:00:00Z",
								"2024-01-01T00:02:00Z",
							},
							Values: []int{5, 20},
							Total:  25,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(1))
			Expect(response.HitsArr[0].Total).To(Equal(50)) // 25 + 25

			// Should have 3 unique timestamps
			Expect(response.HitsArr[0].Timestamps).To(HaveLen(3))

			// Build a map to check timestamp values
			tsMap := make(map[string]int)
			for i, ts := range response.HitsArr[0].Timestamps {
				tsMap[ts] = response.HitsArr[0].Values[i]
			}

			// Check aggregated values
			Expect(tsMap["2024-01-01T00:00:00Z"]).To(Equal(15)) // 10 + 5
			Expect(tsMap["2024-01-01T00:01:00Z"]).To(Equal(15)) // 15 + 0
			Expect(tsMap["2024-01-01T00:02:00Z"]).To(Equal(20)) // 0 + 20
		})
	})

	Context("when merging responses with complex field combinations", func() {
		It("should correctly group by all field labels", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api", "env": "prod"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{100},
							Total:      100,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api", "env": "prod"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{50},
							Total:      50,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{"service": "api", "env": "staging"},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{25},
							Total:      25,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(2))

			// Check that prod environment was aggregated
			for _, hit := range response.HitsArr {
				switch hit.Fields["env"] {
				case "prod":
					Expect(hit.Total).To(Equal(150)) // 100 + 50
				case "staging":
					Expect(hit.Total).To(Equal(25))
				}
			}
		})
	})

	Context("when merging responses with empty fields", func() {
		It("should handle empty field maps", func() {
			responses := []handler.Response{
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{10},
							Total:      10,
						},
					},
				},
				{
					HitsArr: []handler.Hits{
						{
							Fields:     map[string]string{},
							Timestamps: []string{"2024-01-01T00:00:00Z"},
							Values:     []int{20},
							Total:      20,
						},
					},
				},
			}

			result, err := hitsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.Response
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.HitsArr).To(HaveLen(1))
			Expect(response.HitsArr[0].Total).To(Equal(30))
		})
	})
})
