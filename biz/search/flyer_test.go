package bizsearch

import (
	"context"
	"testing"
	"time"

	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"netherrealmstudio.com/aishoppercore/m/db"
)

func TestSearchFlyers(t *testing.T) {

	// Create Elasticsearch client
	esc, err := elasticsearch.NewElasticsearchClient("localhost", "10200")
	require.NoError(t, err)
	require.NotNil(t, esc)

	// Create SearchProductBiz instance
	biz := NewSearchFlyerBiz(esc)

	// Test data
	testFlyers := []db.Flyer{
		{
			ProductName:   "Ham Sandwich",
			Store:         "Grocery Store",
			Brand:         "Fresh Foods",
			Description:   "Fresh ham sandwich with lettuce and tomato",
			ImageURL:      "http://example.com/ham-sandwich.jpg",
			OriginalPrice: 599, // $5.99 in cents
			StartDateTime: time.Now().Add(-24*time.Hour).Unix() * 1000,
			EndDateTime:   time.Now().Add(24*time.Hour).Unix() * 1000,
		},
		{
			ProductName:   "Turkey Ham",
			Store:         "Deli Shop",
			Brand:         "Meat Masters",
			Description:   "Sliced turkey ham",
			ImageURL:      "http://example.com/turkey-ham.jpg",
			OriginalPrice: 499, // $4.99 in cents
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

	// Test cases
	tests := []struct {
		name        string
		productName string
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "search for ham",
			productName: "ham",
			wantCount:   2, // Should find both Ham Sandwich and Turkey Ham
			wantErr:     false,
		},
		{
			name:        "search for non-existent product",
			productName: "xyz123nonexistent",
			wantCount:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Search for products
			flyers, err := biz.SearchFlyers(context.Background(), tt.productName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, flyers)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, flyers)
				assert.Len(t, flyers, tt.wantCount)

				// If we found products, verify their structure
				if len(flyers) > 0 {
					for _, flyer := range flyers {
						assert.NotEmpty(t, flyer.ProductName)
						assert.NotEmpty(t, flyer.Store)
						assert.True(t, flyer.StartDateTime <= time.Now().Unix()*1000)
						assert.True(t, flyer.EndDateTime >= time.Now().Unix()*1000)
					}
				}
			}
		})
	}

	// Cleanup: Delete test index
	err = esc.DeleteIndex(context.Background(), "flyers")
	require.NoError(t, err)
}
