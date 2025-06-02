package bizmatch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kdjuwidja/aishoppercommon/elasticsearch"

	dbmodels "netherealmstudio.com/m/v2/db"
)

// Helper functions for safely extracting values from raw JSON data
func getString(rawData map[string]interface{}, key string) string {
	if val, ok := rawData[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(rawData map[string]interface{}, key string) int {
	if val, ok := rawData[key]; ok && val != nil {
		if num, ok := val.(float64); ok {
			return int(num)
		}
	}
	return 0
}

func getInt64(rawData map[string]interface{}, key string) int64 {
	if val, ok := rawData[key]; ok && val != nil {
		if num, ok := val.(float64); ok {
			return int64(num)
		}
	}
	return 0
}

func getStringArray(rawData map[string]interface{}, key string) []string {
	if val, ok := rawData[key]; ok && val != nil {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, v := range arr {
				if str, ok := v.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}

func (b *MatchShoplistItemsWithFlyerBiz) MatchShoplistItemsWithFlyer(ctx context.Context, shoplistItems []*dbmodels.ShoplistItem) (map[int][]*dbmodels.Flyer, error) {
	now := time.Now().UnixMilli()

	esMultiQuery := elasticsearch.CreateMQuery()

	for _, shoplistItem := range shoplistItems {
		esQuery := elasticsearch.CreateESQueryStr("flyers", newMatchQueryStr(shoplistItem.BrandName, now, shoplistItem.ItemName))
		esMultiQuery.AddQuery(esQuery)
	}

	esMultiQuery.PrintQuery("flyers")

	results, err := b.esc.SearchDocumentsWithMQuery(ctx, "flyers", esMultiQuery)
	if err != nil {
		return nil, err
	}

	itemToFlyersMap := make(map[int][]*dbmodels.Flyer)
	for i, resultSet := range results {
		flyers := make([]*dbmodels.Flyer, 0)
		for _, result := range resultSet {
			// First unmarshal into a map to handle null fields
			var rawData map[string]interface{}
			if err := json.Unmarshal(result, &rawData); err != nil {
				return nil, err
			}

			// Create a new Flyer and populate fields individually
			flyer := &dbmodels.Flyer{}

			// Populate fields using the helper functions
			flyer.Store = getString(rawData, "store")
			flyer.Brand = getString(rawData, "brand")
			flyer.ProductName = getString(rawData, "product_name")
			flyer.Description = getString(rawData, "description")
			flyer.DisclaimerText = getString(rawData, "disclaimer_text")
			flyer.ImageURL = getString(rawData, "image_url")
			flyer.Images = getStringArray(rawData, "images")
			flyer.OriginalPrice = getInt(rawData, "original_price")
			flyer.PrePriceText = getString(rawData, "pre_price_text")
			flyer.PriceText = getString(rawData, "price_text")
			flyer.PostPriceText = getString(rawData, "post_price_text")
			flyer.StartDateTime = getInt64(rawData, "start_date")
			flyer.EndDateTime = getInt64(rawData, "end_date")

			flyers = append(flyers, flyer)
		}
		itemToFlyersMap[shoplistItems[i].ID] = flyers
	}

	return itemToFlyersMap, nil
}
