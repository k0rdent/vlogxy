package handler

import (
	"strconv"

	"github.com/k0rdent/vlogxy/internal/parser"
	log "github.com/sirupsen/logrus"
)

const (
	facetFieldName  = "field_name"
	facetFieldValue = "field_value"
	facetHits       = "hits"
)

type groupKey struct {
	fieldName  string
	fieldValue string
}

// mergeFacetPipe merges rows produced by a facets pipe. Each row contains
// field_name, field_value, and hits. Rows with the same (field_name, field_value)
// pair are collapsed by summing their hits.
func mergeFacetPipe(_ *parser.Pipe, rows []map[string]string) ([]map[string]string, error) {
	sums := make(map[groupKey]int64)
	order := make([]groupKey, 0, len(rows))

	for _, row := range rows {
		key := groupKey{
			fieldName:  row[facetFieldName],
			fieldValue: row[facetFieldValue],
		}

		hits, err := strconv.ParseInt(row[facetHits], 10, 64)
		if err != nil {
			log.Errorf("facets merger: invalid hits value %q: %v", row[facetHits], err)
			continue
		}

		if _, seen := sums[key]; !seen {
			order = append(order, key)
		}
		sums[key] += hits
	}

	result := make([]map[string]string, 0, len(order))
	for _, key := range order {
		result = append(result, map[string]string{
			facetFieldName:  key.fieldName,
			facetFieldValue: key.fieldValue,
			facetHits:       strconv.FormatInt(sums[key], 10),
		})
	}
	return result, nil
}
