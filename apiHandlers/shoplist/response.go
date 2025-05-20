package apiHandlersshoplist

type ShoplistResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Owner struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
	} `json:"owner"`
	Items []struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
		IsBought  bool   `json:"is_bought"`
	} `json:"items"`
}
