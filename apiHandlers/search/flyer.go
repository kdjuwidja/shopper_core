package apiHandlerssearch

import (
	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"netherealmstudio.com/m/v2/apiHandlers"
	"netherealmstudio.com/m/v2/db"
)

func (h *SearchHandler) SearchFlyers(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	searchName := c.Query("searchName")

	if searchName == "" {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "searchName")
		return
	}

	flyers, err := h.searchFlyerBiz.SearchFlyers(c.Request.Context(), searchName)
	if err != nil {
		logger.Errorf("SearchFlyers: Failed to search flyers. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
	}

	if len(flyers) == 0 {
		h.responseFactory.CreateOKResponse(c, gin.H{"flyers": []db.Flyer{}})
		return
	}

	h.responseFactory.CreateOKResponse(c, gin.H{"flyers": flyers})
}
