package bizmatch

import "fmt"

func newMatchQueryStrWithBrand(brand string, startDate, endDate int64, productName string) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"match_phrase":{"brand":"%s"}},{"range":{"start_date":{"lte":%d}}},{"range":{"end_date":{"gte":%d}}}],"should":[{"match":{"product_name":{"query":"%s","fuzziness":"AUTO"}}}],"minimum_should_match":1}}}`, brand, startDate, endDate, productName)
}

func newMatchQueryStrWithoutBrand(startDate, endDate int64, productName string) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"range":{"start_date":{"lte":%d}}},{"range":{"end_date":{"gte":%d}}}],"should":[{"match":{"product_name":{"query":"%s","fuzziness":"AUTO"}}}],"minimum_should_match":1}}}`, startDate, endDate, productName)
}

func newMatchQueryStr(brand string, now int64, productName string) string {
	if brand == "" {
		return newMatchQueryStrWithoutBrand(now, now, productName)
	}
	return newMatchQueryStrWithBrand(brand, now, now, productName)
}
