package bizshoplist

import (
	"context"
	"sort"

	bizmodels "netherealmstudio.com/m/v2/biz"
	"netherealmstudio.com/m/v2/db"
)

// checkShoplistMembership checks if a user is a member of a shoplist
func (b *ShoplistBiz) checkShoplistMembershipFromDB(ctx context.Context, userID string, shoplistID int) bool {
	var member db.ShoplistMember
	err := b.dbPool.GetDB().WithContext(ctx).Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).First(&member).Error
	return err == nil
}

// helper function to get shoplist and members relationship and transform the data into struct for easier use
func (b *ShoplistBiz) GetShoplistWithMembers(ctx context.Context, shoplistID int) (*ShoplistData, *ShoplistError) {
	rows, err := b.dbPool.GetDB().WithContext(ctx).Raw(`SELECT shop_list_id, owner_id, member_id, nickname as member_nickname FROM users RIGHT JOIN
		(SELECT shoplists.id as shop_list_id, shoplists.owner_id as owner_id, shoplist_members.member_id as member_id from shoplists 
		LEFT JOIN shoplist_members ON shoplists.id = shoplist_members.shop_list_id WHERE shoplists.id = ?) as tbl1 ON tbl1.member_id = users.id`, shoplistID).Rows()

	if err != nil {
		return nil, NewShoplistError(ShoplistNotFound, err.Error())
	}
	defer rows.Close()

	type QueryResult struct {
		ShopListID int    `json:"shop_list_id" gorm:"column:shop_list_id"`
		OwnerID    string `json:"owner_id" gorm:"column:owner_id"`
		MemberID   string `json:"member_id" gorm:"column:member_id"`
		Nickname   string `json:"nickname" gorm:"column:nickname"`
	}

	hasRows := false

	// Transform the response data into the operable shoplist data
	var shopListData ShoplistData
	for rows.Next() {
		hasRows = true
		var queryShoplist QueryResult
		err := rows.Scan(&queryShoplist.ShopListID, &queryShoplist.OwnerID, &queryShoplist.MemberID, &queryShoplist.Nickname)
		if err != nil {
			return nil, NewShoplistError(ShoplistNotFound, err.Error())
		}

		if shopListData.ShopListID == 0 {
			shopListData.ShopListID = queryShoplist.ShopListID
			shopListData.OwnerID = queryShoplist.OwnerID
			shopListData.Members = make(map[string]struct {
				MemberID string
				Nickname string
			})
		}

		if _, exists := shopListData.Members[queryShoplist.MemberID]; !exists {
			shopListData.Members[queryShoplist.MemberID] = struct {
				MemberID string
				Nickname string
			}{MemberID: queryShoplist.MemberID, Nickname: queryShoplist.Nickname}
		}
	}

	if !hasRows {
		return nil, NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	return &shopListData, nil
}

// CreateShoplist creates a new shoplist and adds the owner as a member
func (b *ShoplistBiz) CreateShoplist(ctx context.Context, ownerID string, name string) *ShoplistError {
	// Create new shoplist
	shoplist := db.Shoplist{
		OwnerID: ownerID,
		Name:    name,
	}

	// Start a new transaction
	tx := b.dbPool.GetDB().WithContext(ctx).Begin()

	// Save to database
	if err := tx.Create(&shoplist).Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		return NewShoplistError(ShoplistFailedToCreate, err.Error())
	}

	member := db.ShoplistMember{
		ShopListID: shoplist.ID,
		MemberID:   ownerID,
	}

	// Add the owner as a member of the shoplist
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		return NewShoplistError(ShoplistFailedToCreate, err.Error())
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return NewShoplistError(ShoplistFailedToCreate, err.Error())
	}

	return nil
}

// UpdateShoplist updates a shoplist's name
func (b *ShoplistBiz) UpdateShoplist(ctx context.Context, userID string, shoplistID int, name string) *ShoplistError {

	shoplistMembership, err := b.GetShoplistWithMembers(ctx, shoplistID)
	if err != nil {
		return err
	}

	// Check if user is a member
	if _, exists := shoplistMembership.Members[userID]; !exists {
		return NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	// Check if user is the owner
	if shoplistMembership.OwnerID != userID {
		return NewShoplistError(ShoplistNotOwned, "User is not the owner of the shoplist")
	}

	// Update shoplist name
	if err := b.dbPool.GetDB().WithContext(ctx).Model(&db.Shoplist{}).Where("id = ?", shoplistID).Update("name", name).Error; err != nil {
		return NewShoplistError(ShoplistFailedToUpdate, err.Error())
	}

	return nil
}

func (b *ShoplistBiz) GetAllShoplistAndItemsForUser(ctx context.Context, userID string) ([]*bizmodels.Shoplist, *ShoplistError) {
	type QueryResult struct {
		ShopListID    int     `gorm:"column:shop_list_id"`
		ShopListName  string  `gorm:"column:shop_list_name"`
		MemberID      string  `gorm:"column:member_id"`
		ItemID        *int    `gorm:"column:item_id"`
		ItemName      *string `gorm:"column:item_name"`
		BrandName     *string `gorm:"column:brand_name"`
		ExtraInfo     *string `gorm:"column:extra_info"`
		IsBought      *bool   `gorm:"column:is_bought"`
		OwnerID       string  `gorm:"column:owner_id"`
		OwnerNickname string  `gorm:"column:owner_nickname"`
	}

	var results []QueryResult
	err := b.dbPool.GetDB().WithContext(ctx).Raw(`
		SELECT tbl2.shop_list_id as shop_list_id, shop_list_name, member_id, shoplist_items.id as item_id, item_name, brand_name, extra_info, is_bought, owner_id, owner_nickname 
		FROM (
			SELECT shop_list_id, owner_id, nickname as owner_nickname, shop_list_name, member_id 
			FROM (
				SELECT shoplists.id as shop_list_id, owner_id, name as shop_list_name, member_id 
				FROM shoplist_members 
				LEFT JOIN shoplists ON shoplist_members.shop_list_id = shoplists.id 
				WHERE member_id = ?
			) as tbl1 
			LEFT JOIN users on owner_id = users.id
		) as tbl2
		LEFT JOIN shoplist_items on tbl2.shop_list_id = shoplist_items.shop_list_id`, userID).Scan(&results).Error

	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to get shoplist items.")
	}

	if len(results) == 0 {
		return nil, NewShoplistError(ShoplistNotFound, "No shoplists found.")
	}

	// Create a map to store unique shoplists
	shoplistMap := make(map[int]*bizmodels.Shoplist)

	// Process results
	for _, r := range results {
		// Add shoplist if not exists
		if _, exists := shoplistMap[r.ShopListID]; !exists {
			shoplistMap[r.ShopListID] = &bizmodels.Shoplist{
				ID:            r.ShopListID,
				Name:          r.ShopListName,
				OwnerID:       r.OwnerID,
				OwnerNickname: r.OwnerNickname,
				Items:         make([]bizmodels.ShoplistItem, 0),
			}
		}

		// Add item if exists
		if r.ItemID != nil {
			item := bizmodels.ShoplistItem{
				ID:         *r.ItemID,
				ShopListID: r.ShopListID,
				ItemName:   *r.ItemName,
				BrandName:  *r.BrandName,
				ExtraInfo:  *r.ExtraInfo,
				IsBought:   *r.IsBought,
			}
			shoplistMap[r.ShopListID].Items = append(shoplistMap[r.ShopListID].Items, item)
		}
	}

	// Convert map to slice
	shoplists := make([]*bizmodels.Shoplist, 0, len(shoplistMap))
	for _, shoplist := range shoplistMap {
		shoplists = append(shoplists, shoplist)
	}

	// Sort shoplists by ID
	sort.Slice(shoplists, func(i, j int) bool {
		return shoplists[i].ID < shoplists[j].ID
	})

	return shoplists, nil
}

func (b *ShoplistBiz) GetShoplistAndItems(ctx context.Context, userID string, shoplistID int) (*bizmodels.Shoplist, *ShoplistError) {
	type QueryResult struct {
		ShopListID    int     `gorm:"column:shop_list_id"`
		ShopListName  string  `gorm:"column:shop_list_name"`
		MemberID      string  `gorm:"column:member_id"`
		ItemID        *int    `gorm:"column:item_id"`
		ItemName      *string `gorm:"column:item_name"`
		BrandName     *string `gorm:"column:brand_name"`
		ExtraInfo     *string `gorm:"column:extra_info"`
		IsBought      *bool   `gorm:"column:is_bought"`
		OwnerID       string  `gorm:"column:owner_id"`
		OwnerNickname string  `gorm:"column:owner_nickname"`
	}

	var results []QueryResult
	err := b.dbPool.GetDB().WithContext(ctx).Raw(`
		SELECT tbl2.shop_list_id as shop_list_id, shop_list_name, member_id, shoplist_items.id as item_id, item_name, brand_name, extra_info, is_bought, owner_id, owner_nickname 
		FROM (
			SELECT shop_list_id, owner_id, nickname as owner_nickname, shop_list_name, member_id 
			FROM (
				SELECT shoplists.id as shop_list_id, owner_id, name as shop_list_name, member_id 
				FROM shoplist_members 
				LEFT JOIN shoplists ON shoplist_members.shop_list_id = shoplists.id 
				WHERE member_id = ? and shoplists.id = ?
			) as tbl1 
			LEFT JOIN users on owner_id = users.id
		) as tbl2
		LEFT JOIN shoplist_items on tbl2.shop_list_id = shoplist_items.shop_list_id`, userID, shoplistID).Scan(&results).Error

	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to get shoplist items.")
	}

	if len(results) == 0 {
		return nil, NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	// Create shoplist from first row (all rows have same shoplist info)
	shoplist := &bizmodels.Shoplist{
		ID:            results[0].ShopListID,
		Name:          results[0].ShopListName,
		OwnerID:       results[0].OwnerID,
		OwnerNickname: results[0].OwnerNickname,
		Items:         make([]bizmodels.ShoplistItem, 0),
	}

	// Add items to shoplist
	for _, r := range results {
		if r.ItemID != nil {
			shoplist.Items = append(shoplist.Items, bizmodels.ShoplistItem{
				ID:         *r.ItemID,
				ShopListID: r.ShopListID,
				ItemName:   *r.ItemName,
				BrandName:  *r.BrandName,
				ExtraInfo:  *r.ExtraInfo,
				IsBought:   *r.IsBought,
			})
		}
	}

	return shoplist, nil
}
