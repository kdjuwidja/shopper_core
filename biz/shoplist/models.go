package bizshoplist

type ShoplistItem struct {
	ID         int
	ShopListID int
	ItemName   string
	BrandName  string
	ExtraInfo  string
	Thumbnail  string
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
}
