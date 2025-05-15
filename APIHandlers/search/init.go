package apiHandlerssearch

import (
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizsearch "netherealmstudio.com/m/v2/biz/search"
)

type SearchHandler struct {
	searchFlyerBiz  *bizsearch.SearchFlyerBiz
	responseFactory apiHandlers.ResponseFactory
}

func InitializeSearchHandler(esc elasticsearch.ElasticsearchClient, responseFactory apiHandlers.ResponseFactory) *SearchHandler {
	return &SearchHandler{
		searchFlyerBiz:  bizsearch.NewSearchFlyerBiz(&esc),
		responseFactory: responseFactory,
	}
}
