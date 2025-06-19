package bizsearch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"netherealmstudio.com/m/v2/db"
)

type SearchFlyerBiz struct {
	esc *elasticsearch.ElasticsearchClient
}

func NewSearchFlyerBiz(esc *elasticsearch.ElasticsearchClient) *SearchFlyerBiz {
	return &SearchFlyerBiz{
		esc: esc,
	}
}

func (s *SearchFlyerBiz) SearchFlyers(ctx context.Context, product_name string) ([]*db.Flyer, error) {
	startOfToday := time.Now().Truncate(24 * time.Hour).Unix()
	endOfToday := startOfToday + 86400
	esQuery := elasticsearch.CreateESQueryStr("products", newSearchQueryStr(product_name, startOfToday, endOfToday))
	results, err := s.esc.SearchDocuments(ctx, esQuery)
	if err != nil {
		return nil, err
	}

	flyers := make([]*db.Flyer, len(results))
	for i, result := range results {
		var flyer db.Flyer
		if err := json.Unmarshal(result, &flyer); err != nil {
			return nil, err
		}
		flyers[i] = &flyer
	}

	return flyers, nil
}
