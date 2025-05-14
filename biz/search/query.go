package bizsearch

import (
	"fmt"
)

func newSearchQueryStr(productName string, startDate, endDate int64) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"match":{"product_name":"%s"}},{"bool":{"should":[{"range":{"start_date":{"gte":%d}}},{"range":{"end_date":{"lte":%d}}}]}}]}}}`, productName, startDate, endDate)
}
