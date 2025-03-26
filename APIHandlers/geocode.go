package apiHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func GetLatLngByAddress(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address parameter is required",
		})
		return
	}

	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Google Maps API key not configured",
		})
		return
	}

	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s",
		url.QueryEscape(address), apiKey)

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to call Google Maps API",
		})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			FormattedAddress string `json:"formatted_address"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse Google Maps API response",
		})
		return
	}

	if len(result.Results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No results found for this address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lat":               result.Results[0].Geometry.Location.Lat,
		"lng":               result.Results[0].Geometry.Location.Lng,
		"formatted_address": result.Results[0].FormattedAddress,
	})
}
