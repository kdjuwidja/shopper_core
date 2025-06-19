package bizmatch

import "fmt"

func newMatchQueryStrWithBrand(brand string, startDate, endDate int64, productName string) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"match_phrase":{"brand_name":"%s"}},{"range":{"valid_from_timestamp":{"lte":%d}}},{"range":{"valid_to_timestamp":{"gte":%d}}}],"should":[{"match_phrase":{"product_name":{"query":"%s","slop":1}}}],"minimum_should_match":1}}}`, brand, startDate, endDate, productName)
}

func newMatchQueryStrWithoutBrand(startDate, endDate int64, productName string) string {
	return fmt.Sprintf(`{"query":{"bool":{"must":[{"range":{"valid_from_timestamp":{"lte":%d}}},{"range":{"valid_to_timestamp":{"gte":%d}}}],"should":[{"match_phrase":{"product_name":{"query":"%s","slop":1}}}],"minimum_should_match":1}}}`, startDate, endDate, productName)
}

func newMatchQueryStr(brand string, now int64, productName string) string {
	if brand == "" {
		return newMatchQueryStrWithoutBrand(now, now, productName)
	}
	return newMatchQueryStrWithBrand(brand, now, now, productName)
}
