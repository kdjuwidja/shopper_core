package apiHandlersmatch

import (
	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"netherealmstudio.com/m/v2/apiHandlers"
)

func (h *MatchHandler) MatchShoplistItemsWithFlyer(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	var req struct {
		ItemIDs []int `json:"item_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInvalidRequestBody)
		return
	}

	if len(req.ItemIDs) == 0 {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInvalidRequestBody)
		return
	}

	shoplistItems, slerr := h.matchFlyerBiz.GetShoplistItems(c.Request.Context(), userID, req.ItemIDs)
	if slerr != nil {
		logger.Debugf("Error getting shoplist items: %v", slerr)
		h.responseFactory.CreateOKResponse(c, map[string]interface{}{})
		return
	}

	flyers, mferr := h.matchFlyerBiz.MatchShoplistItemsWithFlyer(c.Request.Context(), shoplistItems)
	if mferr != nil {
		logger.Debugf("Error matching shoplist items with flyers: %v", mferr)
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Transform the response to the requested format
	response := make(map[int][]map[string]interface{})
	for itemID, itemFlyers := range flyers {
		flyerDetails := make([]map[string]interface{}, 0, len(itemFlyers))
		for _, flyer := range itemFlyers {
			flyerDetails = append(flyerDetails, map[string]interface{}{
				"store":           flyer.Store,
				"brand":           flyer.Brand,
				"product_name":    flyer.ProductName,
				"description":     flyer.Description,
				"disclaimer_text": flyer.DisclaimerText,
				"image_url":       flyer.ImageURL,
				"images":          flyer.Images,
				"original_price":  flyer.OriginalPrice,
				"pre_price_text":  flyer.PrePriceText,
				"price_text":      flyer.PriceText,
				"post_price_text": flyer.PostPriceText,
				"start_date":      flyer.StartDateTime,
				"end_date":        flyer.EndDateTime,
			})
		}
		if len(flyerDetails) > 0 {
			response[itemID] = flyerDetails
		}
	}

	h.responseFactory.CreateOKResponse(c, response)
}
