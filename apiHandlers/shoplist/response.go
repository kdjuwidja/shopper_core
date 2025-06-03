package apiHandlersshoplist

type ShoplistResponse struct {
	ID    int            `json:"id"`
	Name  string         `json:"name"`
	Owner OwnerResponse  `json:"owner"`
	Items []ItemResponse `json:"items"`
}

type OwnerResponse struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type ItemResponse struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	BrandName string          `json:"brand_name"`
	ExtraInfo string          `json:"extra_info"`
	IsBought  bool            `json:"is_bought"`
	Flyer     []FlyerResponse `json:"flyer"`
}

type FlyerResponse struct {
	Store         string `json:"store"`
	Brand         string `json:"brand"`
	StartDate     int64  `json:"start_date"`
	EndDate       int64  `json:"end_date"`
	ProductName   string `json:"product_name"`
	Description   string `json:"description"`
	OriginalPrice int64  `json:"original_price"`
	PrePriceText  string `json:"pre_price_text"`
	PriceText     string `json:"price_text"`
	PostPriceText string `json:"post_price_text"`
}
