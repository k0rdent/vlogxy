package handler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k0rdent/vlogxy/internal/handler"
)

var _ = Describe("FieldValuesQuery Merge method tests", func() {
	var fieldValuesQuery *handler.FieldValuesQuery

	BeforeEach(func() {
		fieldValuesQuery = handler.NewFieldValuesQuery().(*handler.FieldValuesQuery)
	})

	Context("when merging empty responses", func() {
		It("should return empty values array", func() {
			result, err := fieldValuesQuery.Merge([]handler.FieldValuesResponse{})
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Values).To(BeEmpty())
		})
	})

	Context("when merging single response", func() {
		It("should return the same values", func() {
			responses := []handler.FieldValuesResponse{
				{
					Values: []handler.Value{
						{Value: "value1", Hits: 10},
						{Value: "value2", Hits: 20},
					},
				},
			}

			result, err := fieldValuesQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Values).To(HaveLen(2))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value1", Hits: 10}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value2", Hits: 20}))
		})
	})

	Context("when merging multiple responses with same values", func() {
		It("should aggregate hits for the same value", func() {
			responses := []handler.FieldValuesResponse{
				{
					Values: []handler.Value{
						{Value: "value1", Hits: 10},
						{Value: "value2", Hits: 20},
					},
				},
				{
					Values: []handler.Value{
						{Value: "value1", Hits: 15},
						{Value: "value3", Hits: 5},
					},
				},
			}

			result, err := fieldValuesQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Values).To(HaveLen(3))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value1", Hits: 25}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value2", Hits: 20}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value3", Hits: 5}))
		})
	})

	Context("when merging multiple responses with distinct values", func() {
		It("should include all unique values", func() {
			responses := []handler.FieldValuesResponse{
				{
					Values: []handler.Value{
						{Value: "apple", Hits: 5},
						{Value: "banana", Hits: 3},
					},
				},
				{
					Values: []handler.Value{
						{Value: "cherry", Hits: 7},
						{Value: "date", Hits: 2},
					},
				},
			}

			result, err := fieldValuesQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Values).To(HaveLen(4))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "apple", Hits: 5}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "banana", Hits: 3}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "cherry", Hits: 7}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "date", Hits: 2}))
		})
	})

	Context("when merging responses with zero hits", func() {
		It("should handle zero hits correctly", func() {
			responses := []handler.FieldValuesResponse{
				{
					Values: []handler.Value{
						{Value: "value1", Hits: 0},
					},
				},
				{
					Values: []handler.Value{
						{Value: "value1", Hits: 10},
					},
				},
			}

			result, err := fieldValuesQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Values).To(HaveLen(1))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "value1", Hits: 10}))
		})
	})

	Context("when merging many responses", func() {
		It("should correctly aggregate across all responses", func() {
			responses := []handler.FieldValuesResponse{
				{Values: []handler.Value{{Value: "common", Hits: 1}}},
				{Values: []handler.Value{{Value: "common", Hits: 2}}},
				{Values: []handler.Value{{Value: "common", Hits: 3}}},
				{Values: []handler.Value{{Value: "common", Hits: 4}}},
				{Values: []handler.Value{{Value: "rare", Hits: 1}}},
			}

			result, err := fieldValuesQuery.Merge(responses)
			Expect(err).NotTo(HaveOccurred())

			var response handler.FieldValuesResponse
			err = json.Unmarshal(result, &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Values).To(HaveLen(2))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "common", Hits: 10}))
			Expect(response.Values).To(ContainElement(handler.Value{Value: "rare", Hits: 1}))
		})
	})
})
