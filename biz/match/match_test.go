package bizmatch

import (
	"context"
	"testing"
	"time"

	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/stretchr/testify/assert"
	"netherealmstudio.com/m/v2/db"
)

func setupMatchTestData(t *testing.T, esc *elasticsearch.ElasticsearchClient) {
	ctx := context.Background()
	index := "flyers"

	// Delete index if it exists
	_ = esc.DeleteIndex(ctx, index)

	// Create test documents
	testDocs := []map[string]interface{}{
		{
			"product_name": "Test Item",
			"brand":        "Test Brand",
			"start_date":   time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Generic Item",
			"brand":        "",
			"start_date":   time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Item 1",
			"brand":        "Brand A",
			"start_date":   time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Item 2",
			"brand":        "Brand B",
			"start_date":   time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Item 3",
			"brand":        "Brand C",
			"start_date":   time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(24 * time.Hour).UnixMilli(),
		},
		// Expired items
		{
			"product_name": "Expired Item 1",
			"brand":        "Brand A",
			"start_date":   time.Now().Add(-14 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(-7 * 24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Expired Item 2",
			"brand":        "Brand B",
			"start_date":   time.Now().Add(-21 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(-14 * 24 * time.Hour).UnixMilli(),
		},
		// Future items
		{
			"product_name": "Future Item 1",
			"brand":        "Brand A",
			"start_date":   time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(14 * 24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Future Item 2",
			"brand":        "Brand B",
			"start_date":   time.Now().Add(14 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(21 * 24 * time.Hour).UnixMilli(),
		},
		// Items with overlapping dates
		{
			"product_name": "Overlapping Item 1",
			"brand":        "Brand C",
			"start_date":   time.Now().Add(-14 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
		},
		{
			"product_name": "Overlapping Item 2",
			"brand":        "Brand D",
			"start_date":   time.Now().Add(-21 * 24 * time.Hour).UnixMilli(),
			"end_date":     time.Now().Add(14 * 24 * time.Hour).UnixMilli(),
		},
	}

	// Index test documents
	for _, doc := range testDocs {
		err := esc.IndexDocument(ctx, index, doc)
		if err != nil {
			t.Fatalf("Failed to index test document: %v", err)
		}
	}

	// Give Elasticsearch time to index the documents
	time.Sleep(1 * time.Second)
}

func TestMatchShoplistItemsWithFlyer(t *testing.T) {
	// Create real Elasticsearch client
	esc, err := elasticsearch.NewElasticsearchClient("localhost", "10200")
	if err != nil {
		t.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Setup test data
	setupMatchTestData(t, esc)

	// Create business logic instance
	biz := NewMatchShoplistItemsWithFlyerBiz(esc, nil)

	tests := []struct {
		name            string
		shoplistItems   []*db.ShoplistItem
		expectedError   bool
		expectedMatches int
	}{
		{
			name: "single item with brand",
			shoplistItems: []*db.ShoplistItem{
				{
					ID:        1,
					BrandName: "Test Brand",
					ItemName:  "Test Item",
				},
			},
			expectedError:   false,
			expectedMatches: 1,
		},
		{
			name: "single item without brand",
			shoplistItems: []*db.ShoplistItem{
				{
					ID:        2,
					BrandName: "",
					ItemName:  "Generic Item",
				},
			},
			expectedError:   false,
			expectedMatches: 1,
		},
		{
			name: "multiple items with brands",
			shoplistItems: []*db.ShoplistItem{
				{
					ID:        3,
					BrandName: "Brand A",
					ItemName:  "Item 1",
				},
				{
					ID:        4,
					BrandName: "Brand B",
					ItemName:  "Item 2",
				},
				{
					ID:        5,
					BrandName: "Brand C",
					ItemName:  "Item 3",
				},
			},
			expectedError:   false,
			expectedMatches: 3,
		},
		{
			name: "mixed items with and without brands",
			shoplistItems: []*db.ShoplistItem{
				{
					ID:        6,
					BrandName: "Brand A",
					ItemName:  "Item 1",
				},
				{
					ID:        7,
					BrandName: "",
					ItemName:  "Generic Item",
				},
				{
					ID:        8,
					BrandName: "Brand C",
					ItemName:  "Item 3",
				},
			},
			expectedError:   false,
			expectedMatches: 3,
		},
		{
			name:            "empty shoplist items",
			shoplistItems:   []*db.ShoplistItem{},
			expectedError:   true,
			expectedMatches: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute test
			flyerMap, err := biz.MatchShoplistItemsWithFlyer(context.Background(), tt.shoplistItems)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, flyerMap)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, flyerMap)
				assert.Equal(t, tt.expectedMatches, len(flyerMap), "Number of matched items should match expected")

				// Verify each shoplist item has its corresponding flyers
				for _, item := range tt.shoplistItems {
					flyers, exists := flyerMap[item.ID]
					assert.True(t, exists, "Shoplist item %d should have matching flyers", item.ID)
					assert.NotEmpty(t, flyers, "Shoplist item %d should have at least one matching flyer", item.ID)
				}
			}
		})
	}
}

func TestNewMatchShoplistItemsWithFlyerBiz(t *testing.T) {
	esc, err := elasticsearch.NewElasticsearchClient("localhost", "10200")
	if err != nil {
		t.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	biz := NewMatchShoplistItemsWithFlyerBiz(esc, nil)
	assert.NotNil(t, biz)
	assert.Equal(t, esc, biz.esc)
}
