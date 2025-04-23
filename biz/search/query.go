package bizsearch

// NewSearchQuery creates a new search query for products with date range
func newSearchQuery(productName string, startDate, endDate int64) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"product_name": productName,
						},
					},
					map[string]interface{}{
						"bool": map[string]interface{}{
							"should": []interface{}{
								map[string]interface{}{
									"range": map[string]interface{}{
										"start_date": map[string]interface{}{
											"gte": startDate,
										},
									},
								},
								map[string]interface{}{
									"range": map[string]interface{}{
										"end_date": map[string]interface{}{
											"lte": endDate,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
