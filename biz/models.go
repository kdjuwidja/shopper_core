package biz

type ShoplistItem struct {
	ID         int
	ShopListID int
	ItemName   string
	BrandName  string
	ExtraInfo  string
	IsBought   bool
}

type ShoplistMember struct {
	ID       string
	Nickname string
}

type Shoplist struct {
	ID            int
	Name          string
	OwnerID       string
	OwnerNickname string
	Items         []ShoplistItem
}

type Flyer struct {
	Store          string
	Brand          string
	ProductName    string
	Description    string
	DisclaimerText string
	OriginalPrice  int64
	PrePriceText   string
	PriceText      string
	PostPriceText  string
	StartDateTime  int64
	EndDateTime    int64
}
