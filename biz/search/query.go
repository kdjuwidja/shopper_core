package bizsearch

import (
	"fmt"
)

func newSearchQueryStr(productName string, startDate, endDate int64) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"match":{"product_name":"%s"}},{"bool":{"should":[{"range":{"valid_from_timestamp":{"gte":%d}}},{"range":{"valid_to_timestamp":{"lte":%d}}}]}}]}}}`, productName, startDate, endDate)
}
