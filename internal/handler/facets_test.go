package handler

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FacetsQuery Merge method tests", func() {
	var facetsQuery *FacetsQuery

	BeforeEach(func() {
		facetsQuery = NewFacetsQuery().(*FacetsQuery)
	})

	Context("when merging empty responses", func() {
		It("should return empty facets array", func() {
			result, err := facetsQuery.Merge([]FacetsResponse{})
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Facets).To(BeEmpty())
		})
	})

	Context("when merging single response", func() {
		It("should return the same facets", func() {
			responses := []FacetsResponse{
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 50},
								{Value: "404", Hits: 10},
							},
						},
						{
							FieldName: "method",
							Values: []FacetValue{
								{Value: "GET", Hits: 40},
								{Value: "POST", Hits: 20},
							},
						},
					},
				},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(2))

			statusFacet := findFacet(response.Facets, "status")
			Expect(statusFacet).NotTo(BeNil())
			Expect(statusFacet.Values).To(HaveLen(2))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "200", Hits: 50}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "404", Hits: 10}))

			methodFacet := findFacet(response.Facets, "method")
			Expect(methodFacet).NotTo(BeNil())
			Expect(methodFacet.Values).To(HaveLen(2))
			Expect(methodFacet.Values).To(ContainElement(FacetValue{Value: "GET", Hits: 40}))
			Expect(methodFacet.Values).To(ContainElement(FacetValue{Value: "POST", Hits: 20}))
		})
	})

	Context("when merging multiple responses with same field names and values", func() {
		It("should aggregate hits for matching field names and values", func() {
			responses := []FacetsResponse{
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 50},
								{Value: "404", Hits: 10},
							},
						},
					},
				},
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 30},
								{Value: "500", Hits: 5},
							},
						},
					},
				},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(1))
			statusFacet := response.Facets[0]
			Expect(statusFacet.FieldName).To(Equal("status"))
			Expect(statusFacet.Values).To(HaveLen(3))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "200", Hits: 80}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "404", Hits: 10}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "500", Hits: 5}))
		})
	})

	Context("when merging multiple responses with distinct field names", func() {
		It("should include all unique field names", func() {
			responses := []FacetsResponse{
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 50},
							},
						},
					},
				},
				{
					Facets: []Facets{
						{
							FieldName: "method",
							Values: []FacetValue{
								{Value: "GET", Hits: 40},
							},
						},
					},
				},
				{
					Facets: []Facets{
						{
							FieldName: "host",
							Values: []FacetValue{
								{Value: "api.example.com", Hits: 30},
							},
						},
					},
				},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(3))
			fieldNames := extractFieldNames(response.Facets)
			Expect(fieldNames).To(ContainElement("status"))
			Expect(fieldNames).To(ContainElement("method"))
			Expect(fieldNames).To(ContainElement("host"))
		})
	})

	Context("when merging multiple responses with overlapping and distinct values", func() {
		It("should correctly aggregate overlapping values and include distinct ones", func() {
			responses := []FacetsResponse{
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 50},
								{Value: "404", Hits: 10},
								{Value: "500", Hits: 2},
							},
						},
					},
				},
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 45},
								{Value: "201", Hits: 15},
								{Value: "503", Hits: 3},
							},
						},
					},
				},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(1))
			statusFacet := response.Facets[0]
			Expect(statusFacet.Values).To(HaveLen(5))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "200", Hits: 95}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "404", Hits: 10}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "500", Hits: 2}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "201", Hits: 15}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "503", Hits: 3}))
		})
	})

	Context("when merging responses with zero hits", func() {
		It("should handle zero hits correctly", func() {
			responses := []FacetsResponse{
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 0},
							},
						},
					},
				},
				{
					Facets: []Facets{
						{
							FieldName: "status",
							Values: []FacetValue{
								{Value: "200", Hits: 50},
							},
						},
					},
				},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(1))
			statusFacet := response.Facets[0]
			Expect(statusFacet.Values).To(HaveLen(1))
			Expect(statusFacet.Values[0]).To(Equal(FacetValue{Value: "200", Hits: 50}))
		})
	})

	Context("when merging many responses", func() {
		It("should correctly aggregate across all responses", func() {
			responses := []FacetsResponse{
				{Facets: []Facets{{FieldName: "status", Values: []FacetValue{{Value: "200", Hits: 1}}}}},
				{Facets: []Facets{{FieldName: "status", Values: []FacetValue{{Value: "200", Hits: 2}}}}},
				{Facets: []Facets{{FieldName: "status", Values: []FacetValue{{Value: "200", Hits: 3}}}}},
				{Facets: []Facets{{FieldName: "status", Values: []FacetValue{{Value: "200", Hits: 4}}}}},
				{Facets: []Facets{{FieldName: "status", Values: []FacetValue{{Value: "404", Hits: 1}}}}},
			}

			result, err := facetsQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response FacetsResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Facets).To(HaveLen(1))
			statusFacet := response.Facets[0]
			Expect(statusFacet.Values).To(HaveLen(2))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "200", Hits: 10}))
			Expect(statusFacet.Values).To(ContainElement(FacetValue{Value: "404", Hits: 1}))
		})
	})
})

func findFacet(facets []Facets, fieldName string) *Facets {
	for i := range facets {
		if facets[i].FieldName == fieldName {
			return &facets[i]
		}
	}
	return nil
}

func extractFieldNames(facets []Facets) []string {
	names := make([]string, 0, len(facets))
	for _, facet := range facets {
		names = append(names, facet.FieldName)
	}
	return names
}
