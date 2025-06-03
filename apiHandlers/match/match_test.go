package apiHandlersmatch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiHandlers "netherealmstudio.com/m/v2/apiHandlers"
	dbmodel "netherealmstudio.com/m/v2/db"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

const (
	elasticsearchHost = "localhost"
	elasticsearchPort = "10200"
	mysqlUser         = "root"
	mysqlPassword     = "password"
	mysqlHost         = "localhost"
	mysqlPort         = "4306"
	mysqlDB           = "test_db"
)

func setupTestData(t *testing.T, dbPool *db.MySQLConnectionPool, esClient *elasticsearch.ElasticsearchClient) {
	// Delete existing index if it exists
	err := esClient.DeleteIndex(context.Background(), "flyers")
	require.NoError(t, err)

	// Create test users
	users := []*dbmodel.User{
		{
			ID:         "test_user_1",
			Nickname:   "Test User 1",
			PostalCode: "123456",
		},
		{
			ID:         "test_user_2",
			Nickname:   "Test User 2",
			PostalCode: "654321",
		},
	}
	for _, user := range users {
		err := dbPool.GetDB().Create(user).Error
		require.NoError(t, err)
	}

	// Create test shoplists
	shoplists := []*dbmodel.Shoplist{
		{
			OwnerID: users[0].ID,
			Name:    "Weekly Groceries",
		},
		{
			OwnerID: users[0].ID,
			Name:    "Party Supplies",
		},
		{
			OwnerID: users[1].ID,
			Name:    "Office Supplies",
		},
	}
	for _, shoplist := range shoplists {
		err := dbPool.GetDB().Create(shoplist).Error
		require.NoError(t, err)
	}

	// Add test_user_1 as a member to all shoplists
	for _, shoplist := range shoplists {
		member := &dbmodel.ShoplistMember{
			ShopListID: shoplist.ID,
			MemberID:   "test_user_1",
		}
		err := dbPool.GetDB().Create(member).Error
		require.NoError(t, err)
	}

	// Create test shoplist items
	items := []dbmodel.ShoplistItem{
		// Items in Weekly Groceries
		{
			ShopListID: shoplists[0].ID,
			ItemName:   "Milk",
			BrandName:  "Dairy Fresh",
			ExtraInfo:  "2%",
			IsBought:   false,
		},
		{
			ShopListID: shoplists[0].ID,
			ItemName:   "Bread",
			BrandName:  "Wonder",
			ExtraInfo:  "Whole Wheat",
			IsBought:   false,
		},
		// Items in Party Supplies
		{
			ShopListID: shoplists[1].ID,
			ItemName:   "Chips",
			BrandName:  "Lays",
			ExtraInfo:  "Classic",
			IsBought:   false,
		},
		{
			ShopListID: shoplists[1].ID,
			ItemName:   "Soda",
			BrandName:  "Coca-Cola",
			ExtraInfo:  "2L",
			IsBought:   false,
		},
		// Items in Office Supplies
		{
			ShopListID: shoplists[2].ID,
			ItemName:   "Paper",
			BrandName:  "HP",
			ExtraInfo:  "A4",
			IsBought:   false,
		},
		// Items shared between shoplists
		{
			ShopListID: shoplists[0].ID,
			ItemName:   "Water",
			BrandName:  "Dasani",
			ExtraInfo:  "24-pack",
			IsBought:   false,
		},
		{
			ShopListID: shoplists[1].ID,
			ItemName:   "Water",
			BrandName:  "Dasani",
			ExtraInfo:  "24-pack",
			IsBought:   false,
		},
	}
	for _, item := range items {
		err := dbPool.GetDB().Create(&item).Error
		require.NoError(t, err)
	}

	// Create test flyer data in Elasticsearch
	flyers := []dbmodel.Flyer{
		{
			Store:          "Walmart",
			Brand:          "Dairy Fresh",
			ProductName:    "Milk",
			Description:    "Fresh 2% Milk",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/milk.jpg",
			Images:         []string{"http://example.com/milk1.jpg", "http://example.com/milk2.jpg"},
			OriginalPrice:  "4.99",
			PrePriceText:   "Was",
			PriceText:      "$3.99",
			PostPriceText:  "each",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			Store:          "Target",
			Brand:          "Wonder",
			ProductName:    "Bread",
			Description:    "Whole Wheat Bread",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/bread.jpg",
			Images:         []string{"http://example.com/bread1.jpg", "http://example.com/bread2.jpg"},
			OriginalPrice:  "3.99",
			PrePriceText:   "Was",
			PriceText:      "$2.99",
			PostPriceText:  "each",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			Store:          "Costco",
			Brand:          "Lays",
			ProductName:    "Chips",
			Description:    "Classic Potato Chips",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/chips.jpg",
			Images:         []string{"http://example.com/chips1.jpg", "http://example.com/chips2.jpg"},
			OriginalPrice:  "2.99",
			PrePriceText:   "Was",
			PriceText:      "$2.49",
			PostPriceText:  "each",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			Store:          "Safeway",
			Brand:          "Coca-Cola",
			ProductName:    "Soda",
			Description:    "2L Bottle",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/soda.jpg",
			Images:         []string{"http://example.com/soda1.jpg", "http://example.com/soda2.jpg"},
			OriginalPrice:  "1.99",
			PrePriceText:   "Was",
			PriceText:      "$1.99",
			PostPriceText:  "each",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			Store:          "Office Depot",
			Brand:          "HP",
			ProductName:    "Paper",
			Description:    "A4 Copy Paper",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/paper.jpg",
			Images:         []string{"http://example.com/paper1.jpg", "http://example.com/paper2.jpg"},
			OriginalPrice:  "5.99",
			PrePriceText:   "Was",
			PriceText:      "$5.99",
			PostPriceText:  "per ream",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			Store:          "Walmart",
			Brand:          "Dasani",
			ProductName:    "Water",
			Description:    "24-pack Bottled Water",
			DisclaimerText: "While supplies last",
			ImageURL:       "http://example.com/water.jpg",
			Images:         []string{"http://example.com/water1.jpg", "http://example.com/water2.jpg"},
			OriginalPrice:  "3.99",
			PrePriceText:   "Was",
			PriceText:      "$3.99",
			PostPriceText:  "per pack",
			StartDateTime:  time.Now().Add(-3 * 24 * time.Hour).UnixMilli(),
			EndDateTime:    time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
	}

	// Index flyers in Elasticsearch
	for _, flyer := range flyers {
		err := esClient.IndexDocument(context.Background(), "flyers", flyer)
		require.NoError(t, err)
	}

	// Wait for indexing to complete
	time.Sleep(1 * time.Second)
}

func TestMatchShoplistItemsWithFlyer(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewElasticsearchClient(elasticsearchHost, elasticsearchPort)
	require.NoError(t, err)

	// Initialize MySQL connection pool using testutil
	dbPool := testutil.SetupTestEnv(t)
	defer testutil.TeardownTestDB(dbPool)

	// Set up test data
	setupTestData(t, dbPool, esClient)

	// Initialize business logic components
	responseFactory := apiHandlers.Initialize()

	// Initialize handler
	handler := InitializeMatchHandler(
		*esClient,
		*dbPool,
		*responseFactory,
	)

	// Add mock user ID middleware
	router.Use(func(c *gin.Context) {
		c.Set("userID", "test_user_1")
		c.Next()
	})

	router.POST("/match", handler.MatchShoplistItemsWithFlyer)

	tests := []struct {
		name           string
		userID         string
		itemIDs        []int
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "Single item match",
			userID:         "test_user_1",
			itemIDs:        []int{1},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Check Milk match
				milkFlyers, ok := response["1"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(milkFlyers), 0)
				milkFlyer := milkFlyers[0].(map[string]interface{})
				assert.Equal(t, "Walmart", milkFlyer["store"])
				assert.Equal(t, "Dairy Fresh", milkFlyer["brand"])
				assert.Equal(t, "Milk", milkFlyer["product_name"])
			},
		},
		{
			name:           "Multiple items from same shoplist",
			userID:         "test_user_1",
			itemIDs:        []int{1, 2},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Check Milk match
				milkFlyers, ok := response["1"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(milkFlyers), 0)
				milkFlyer := milkFlyers[0].(map[string]interface{})
				assert.Equal(t, "Walmart", milkFlyer["store"])
				assert.Equal(t, "Dairy Fresh", milkFlyer["brand"])
				assert.Equal(t, "Milk", milkFlyer["product_name"])

				// Check Bread match
				breadFlyers, ok := response["2"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(breadFlyers), 0)
				breadFlyer := breadFlyers[0].(map[string]interface{})
				assert.Equal(t, "Target", breadFlyer["store"])
				assert.Equal(t, "Wonder", breadFlyer["brand"])
				assert.Equal(t, "Bread", breadFlyer["product_name"])
			},
		},
		{
			name:           "Items from different shoplists",
			userID:         "test_user_1",
			itemIDs:        []int{1, 3, 5},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Check all items have matches
				assert.Contains(t, response, "1") // Milk
				assert.Contains(t, response, "3") // Chips
				assert.Contains(t, response, "5") // Paper

				// Verify each item has at least one flyer
				for _, itemID := range []string{"1", "3", "5"} {
					flyers, ok := response[itemID].([]interface{})
					require.True(t, ok)
					require.Greater(t, len(flyers), 0)
				}
			},
		},
		{
			name:           "Shared item between shoplists",
			userID:         "test_user_1",
			itemIDs:        []int{6, 7},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Both items should have the same match
				water1Flyers, ok := response["6"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(water1Flyers), 0)
				water1Flyer := water1Flyers[0].(map[string]interface{})
				assert.Equal(t, "Walmart", water1Flyer["store"])
				assert.Equal(t, "Dasani", water1Flyer["brand"])
				assert.Equal(t, "Water", water1Flyer["product_name"])

				water2Flyers, ok := response["7"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(water2Flyers), 0)
				water2Flyer := water2Flyers[0].(map[string]interface{})
				assert.Equal(t, "Walmart", water2Flyer["store"])
				assert.Equal(t, "Dasani", water2Flyer["brand"])
				assert.Equal(t, "Water", water2Flyer["product_name"])
			},
		},
		{
			name:           "Empty item IDs",
			userID:         "test_user_1",
			itemIDs:        []int{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:           "Non-existent item ID",
			userID:         "test_user_1",
			itemIDs:        []int{999},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Empty(t, response)
			},
		},
		{
			name:           "Mixed valid and invalid item IDs",
			userID:         "test_user_1",
			itemIDs:        []int{1, 999, 2, 888},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Check valid items have matches
				assert.Contains(t, response, "1") // Milk
				assert.Contains(t, response, "2") // Bread

				// Verify Milk match
				milkFlyers, ok := response["1"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(milkFlyers), 0)
				milkFlyer := milkFlyers[0].(map[string]interface{})
				assert.Equal(t, "Walmart", milkFlyer["store"])
				assert.Equal(t, "Dairy Fresh", milkFlyer["brand"])
				assert.Equal(t, "Milk", milkFlyer["product_name"])

				// Verify Bread match
				breadFlyers, ok := response["2"].([]interface{})
				require.True(t, ok)
				require.Greater(t, len(breadFlyers), 0)
				breadFlyer := breadFlyers[0].(map[string]interface{})
				assert.Equal(t, "Target", breadFlyer["store"])
				assert.Equal(t, "Wonder", breadFlyer["brand"])
				assert.Equal(t, "Bread", breadFlyer["product_name"])

				// Verify invalid items are not in response
				assert.NotContains(t, response, "999")
				assert.NotContains(t, response, "888")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody := map[string]interface{}{
				"item_ids": tt.itemIDs,
			}
			jsonBody, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/match", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response content
			tt.checkResponse(t, response)
		})
	}
}
