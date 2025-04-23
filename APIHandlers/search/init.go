package apiHandlerssearch

import (
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	bizsearch "netherrealmstudio.com/aishoppercore/m/biz/search"
)

type SearchHandler struct {
	searchFlyerBiz  *bizsearch.SearchFlyerBiz
	responseFactory apiHandlers.ResponseFactory
}

func InitializeSearchHandler(esc *elasticsearch.ElasticsearchClient, responseFactory apiHandlers.ResponseFactory) *SearchHandler {
	return &SearchHandler{
		searchFlyerBiz:  bizsearch.NewSearchFlyerBiz(esc),
		responseFactory: responseFactory,
	}
}
