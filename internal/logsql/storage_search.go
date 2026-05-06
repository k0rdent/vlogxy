package logstorage

func (iff *ifFilter) visitSubqueries(visitFunc func(q *Query)) {
	if iff != nil {
		visitSubqueriesInFilter(iff.f, visitFunc)
	}
}
