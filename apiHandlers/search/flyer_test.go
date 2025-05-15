package apiHandlerssearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizsearch "netherealmstudio.com/m/v2/biz/search"
	"netherealmstudio.com/m/v2/db"
)

func TestSearchFlyers(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create Elasticsearch client
	esc, err := elasticsearch.NewElasticsearchClient("localhost", "10200")
	require.NoError(t, err)
	require.NotNil(t, esc)

	// Cleanup: Delete test index
	esc.DeleteIndex(context.Background(), "flyers")

	// Create test products
	testFlyers := []db.Flyer{
		{
			ProductName:   "Test Product 1",
			Store:         "Test Store",
			Brand:         "Test Brand",
			Description:   "Test Description 1",
			ImageURL:      "http://example.com/test1.jpg",
			OriginalPrice: 1000,
			StartDateTime: time.Now().Add(-24*time.Hour).Unix() * 1000,
			EndDateTime:   time.Now().Add(24*time.Hour).Unix() * 1000,
		},
		{
			ProductName:   "Test Product 2",
			Store:         "Test Store",
			Brand:         "Test Brand",
			Description:   "Test Description 2",
			ImageURL:      "http://example.com/test2.jpg",
			OriginalPrice: 2000,
			StartDateTime: time.Now().Add(-24*time.Hour).Unix() * 1000,
			EndDateTime:   time.Now().Add(24*time.Hour).Unix() * 1000,
		},
	}

	// Index test products
	for _, flyer := range testFlyers {
		err := esc.IndexDocument(context.Background(), "flyers", flyer)
		require.NoError(t, err)
	}

	// Wait for indexing to complete
	time.Sleep(1 * time.Second)

	// Create test cases
	tests := []struct {
		name           string
		searchQuery    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "successful search",
			searchQuery:    "Test",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "no results",
			searchQuery:    "nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "empty query",
			searchQuery:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			router.Use(gin.Recovery())

			// Add mock user ID middleware
			router.Use(func(c *gin.Context) {
				c.Set("userID", "test-user-123")
				c.Next()
			})

			// Create response factory
			responseFactory := apiHandlers.Initialize()

			// Create search handler
			handler := &SearchHandler{
				searchFlyerBiz:  bizsearch.NewSearchFlyerBiz(esc),
				responseFactory: *responseFactory,
			}

			// Register the search endpoint
			router.GET("/search", handler.SearchFlyers)

			// Create request
			req, err := http.NewRequest("GET", "/search?searchName="+tt.searchQuery, nil)
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response struct {
					Flyers []db.Flyer `json:"flyers"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Check number of results
				assert.Len(t, response.Flyers, tt.expectedCount)

				// If we have results, verify their structure
				if tt.expectedCount > 0 {
					for _, flyer := range response.Flyers {
						assert.NotEmpty(t, flyer.ProductName)
						assert.NotEmpty(t, flyer.Store)
						assert.NotEmpty(t, flyer.Brand)
						assert.True(t, flyer.StartDateTime <= time.Now().Unix()*1000)
						assert.True(t, flyer.EndDateTime >= time.Now().Unix()*1000)
					}
				}
			}
		})
	}
}
