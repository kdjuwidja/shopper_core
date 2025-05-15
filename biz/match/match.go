package bizmatch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kdjuwidja/aishoppercommon/elasticsearch"

	dbmodels "netherealmstudio.com/m/v2/db"
)

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
			var flyer dbmodels.Flyer
			if err := json.Unmarshal(result, &flyer); err != nil {
				return nil, err
			}
			flyers = append(flyers, &flyer)
		}
		itemToFlyersMap[shoplistItems[i].ID] = flyers
	}

	return itemToFlyersMap, nil
}
