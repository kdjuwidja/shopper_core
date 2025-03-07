package APIHandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchRequest struct {
	Products  []string `json:"products" binding:"required"`
	Latitude  float64  `json:"latitude" binding:"required"`
	Longitude float64  `json:"longitude" binding:"required"`
	Radius    float64  `json:"radius" binding:"required"` // in kilometers
}

type GooglePlacesRequest struct {
	IncludedTypes       []string `json:"includedTypes"`
	MaxResultCount      int      `json:"maxResultCount"`
	LocationRestriction struct {
		Circle struct {
			Center struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"center"`
			Radius float64 `json:"radius"`
		} `json:"circle"`
	} `json:"locationRestriction"`
}

type Place struct {
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	FormattedAddress string `json:"formattedAddress"`
}

type PlacesResponse struct {
	Places []Place `json:"places"`
}

func callGooglePlacesAPI(request SearchRequest, includedTypes []string, maxResults int) (*PlacesResponse, error) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("google maps API key not configured")
	}

	googleRequest := GooglePlacesRequest{
		IncludedTypes:  includedTypes,
		MaxResultCount: maxResults,
	}
	googleRequest.LocationRestriction.Circle.Center.Latitude = request.Latitude
	googleRequest.LocationRestriction.Circle.Center.Longitude = request.Longitude
	googleRequest.LocationRestriction.Circle.Radius = request.Radius

	jsonData, err := json.Marshal(googleRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://places.googleapis.com/v1/places:searchNearby", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.displayName,places.formattedAddress,places.location")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch places: %v", err)
	}
	defer resp.Body.Close()

	var result PlacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func Recommend(c *gin.Context) {
	var request SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Explicit validation for each field
	if request.Latitude == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude is required"})
		return
	}
	if request.Longitude == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "longitude is required"})
		return
	}
	if request.Radius == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "radius is required"})
		return
	}

	// Prepare request body
	maxResults := 10
	limit := os.Getenv("PLACE_SEARCH_RESULT_LIMIT")
	if limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			maxResults = parsed
		} else {
			fmt.Printf("Error parsing limit: %v\n", err)
		}
	}

	// Get included types from environment
	includedTypes := []string{"supermarket", "convenience_store", "drugstore", "pharmacy", "liquor_store"}
	if types := os.Getenv("PLACE_INCLUDED_TYPES"); types != "" {
		includedTypes = strings.Split(types, ",")
	}

	result, err := callGooglePlacesAPI(request, includedTypes, maxResults)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Extract display names and addresses from results
	type SimplifiedPlace struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	simplifiedResults := make([]SimplifiedPlace, len(result.Places))
	for i, place := range result.Places {
		simplifiedResults[i] = SimplifiedPlace{
			Name:    place.DisplayName.Text,
			Address: place.FormattedAddress,
		}
	}
	// Add the product list to the response
	response := gin.H{
		"products": request.Products,
		"places":   simplifiedResults,
	}

	c.JSON(http.StatusOK, response)
}
