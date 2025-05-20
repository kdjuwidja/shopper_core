package bizshoplist

import (
	"context"
	"sort"

	"netherealmstudio.com/m/v2/db"
)

// checkShoplistMembership checks if a user is a member of a shoplist
func (b *ShoplistBiz) checkShoplistMembershipFromDB(userID string, shoplistID int) bool {
	var member db.ShoplistMember
	err := b.dbPool.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).First(&member).Error
	return err == nil
}

// helper function to get shoplist and members relationship and transform the data into struct for easier use
func (b *ShoplistBiz) GetShoplistWithMembers(shoplistID int) (*ShoplistData, *ShoplistError) {
	rows, err := b.dbPool.GetDB().Raw(`SELECT shop_list_id, owner_id, member_id, nickname as member_nickname FROM users RIGHT JOIN
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
func (b *ShoplistBiz) CreateShoplist(ownerID string, name string) *ShoplistError {
	// Create new shoplist
	shoplist := db.Shoplist{
		OwnerID: ownerID,
		Name:    name,
	}

	// Start a new transaction
	tx := b.dbPool.GetDB().Begin()

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

// GetAllShoplists retrieves all shoplists
func (b *ShoplistBiz) GetAllShoplists(userID string) ([]GetAllShoplistData, *ShoplistError) {
	// Get all shoplists where user is a member
	rows, err := b.dbPool.GetDB().Raw(`
		select shoplists.id as id, shoplists.name as name, users.id as owner_id, users.nickname as owner_nickname from shoplists
		left join users on shoplists.owner_id = users.id
		where shoplists.id in (select shop_list_id from shoplist_members where member_id=?);
	`, userID).Rows()

	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, err.Error())
	}
	defer rows.Close()

	var shoplistData []GetAllShoplistData
	for rows.Next() {
		var r GetAllShoplistData
		err := rows.Scan(&r.ID, &r.Name, &r.OwnerID, &r.OwnerNickname)
		if err != nil {
			return nil, NewShoplistError(ShoplistFailedToProcess, err.Error())
		}
		shoplistData = append(shoplistData, r)
	}

	return shoplistData, nil
}

// GetShoplist retrieves a shoplist by ID
func (b *ShoplistBiz) GetShoplist(userID string, shoplistID int) (*Shoplist, []ShoplistItem, []ShoplistMember, *ShoplistError) {
	if !b.checkShoplistMembershipFromDB(userID, shoplistID) {
		return nil, nil, nil, NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	rows, err := b.dbPool.GetDB().Raw(`SELECT shoplists.id as shop_list_id, shoplists.name as shop_list_name, owner_id, shoplist_items.id as shop_list_item_id, item_name, brand_name, extra_info, is_bought, thumbnail, member_id, member_nickname FROM shoplists
		LEFT JOIN shoplist_items on shoplist_items.shop_list_id = shoplists.id
		LEFT JOIN (SELECT shop_list_id, member_id, nickname as member_nickname from shoplist_members left join users on shoplist_members.member_id = users.id) as tbl1 ON tbl1.shop_list_id = shoplists.id
		where shoplists.id = ?;`, shoplistID).Rows()
	if err != nil {
		return nil, nil, nil, NewShoplistError(ShoplistNotFound, err.Error())
	}
	defer rows.Close()

	var shoplistData Shoplist
	itemMap := make(map[int]ShoplistItem)
	itemIdList := make([]int, 0)
	memberMap := make(map[string]ShoplistMember)
	memberIdList := make([]string, 0)

	for rows.Next() {
		var r GetShoplistData
		err := rows.Scan(&r.ID, &r.Name, &r.OwnerId, &r.ShopListItemID, &r.ShopListItemName, &r.ShopListItemBrandName, &r.ShopListItemExtraInfo, &r.ShopListItemIsBought, &r.ShopListItemThumbnail, &r.ShopListMemberID, &r.ShopListMemberNickname)
		if err != nil {
			return nil, nil, nil, NewShoplistError(ShoplistFailedToProcess, err.Error())
		}

		shoplistData.ID = r.ID
		shoplistData.Name = r.Name
		shoplistData.OwnerID = r.OwnerId

		// Item can be nil if the shoplist is empty
		if r.ShopListItemID != nil {
			if _, exists := itemMap[*r.ShopListItemID]; !exists {
				itemMap[*r.ShopListItemID] = ShoplistItem{
					ID:        *r.ShopListItemID,
					ItemName:  *r.ShopListItemName,
					BrandName: *r.ShopListItemBrandName,
					ExtraInfo: *r.ShopListItemExtraInfo,
					IsBought:  *r.ShopListItemIsBought,
					Thumbnail: *r.ShopListItemThumbnail,
				}
				itemIdList = append(itemIdList, *r.ShopListItemID)
			}
		}

		// At least the owner should be in the member list, so we don't need to check if the member exists
		if _, exists := memberMap[r.ShopListMemberID]; !exists {
			memberMap[r.ShopListMemberID] = ShoplistMember{
				ID:       r.ShopListMemberID,
				Nickname: r.ShopListMemberNickname,
			}
			memberIdList = append(memberIdList, r.ShopListMemberID)
		}
	}

	shoplistData.OwnerNickname = memberMap[shoplistData.OwnerID].Nickname
	// Convert the map to slices
	items := make([]ShoplistItem, 0, len(itemMap))
	sort.Ints(itemIdList)
	for _, itemId := range itemIdList {
		items = append(items, itemMap[itemId])
	}

	members := make([]ShoplistMember, 0, len(memberMap))
	sort.Strings(memberIdList)
	for _, memberId := range memberIdList {
		members = append(members, memberMap[memberId])
	}

	return &shoplistData, items, members, nil
}

// UpdateShoplist updates a shoplist's name
func (b *ShoplistBiz) UpdateShoplist(userID string, shoplistID int, name string) *ShoplistError {

	shoplistMembership, err := b.GetShoplistWithMembers(shoplistID)
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
	if err := b.dbPool.GetDB().Model(&db.Shoplist{}).Where("id = ?", shoplistID).Update("name", name).Error; err != nil {
		return NewShoplistError(ShoplistFailedToUpdate, err.Error())
	}

	return nil
}

func (b *ShoplistBiz) GetAllShoplistAndItemsForUser(ctx context.Context, userID string) ([]*Shoplist, *ShoplistError) {
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
		SELECT tbl2.shop_list_id as shop_list_id, shop_list_name, member_id, item_id, item_name, brand_name, extra_info, is_bought, owner_id, owner_nickname 
		FROM (SELECT * FROM shoplist_members WHERE member_id = ?) as tbl1
		LEFT JOIN (
			SELECT shop_list_id, shop_list_name, item_id, item_name, brand_name, extra_info, is_bought, owner_id, nickname as owner_nickname 
			FROM (
				SELECT shop_list_id as shop_list_id, name as shop_list_name, shoplist_items.id as item_id, item_name, brand_name, extra_info, is_bought, owner_id 
				FROM shoplist_items 
				LEFT JOIN shoplists ON shoplist_items.shop_list_id = shoplists.id
			) as shoplistXshoplistItems 
			LEFT JOIN users ON users.id = shoplistXshoplistItems.owner_id
		) as tbl2
		ON tbl2.shop_list_id = tbl1.shop_list_id`, userID).Scan(&results).Error

	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to get shoplist items.")
	}

	if len(results) == 0 {
		return nil, NewShoplistError(ShoplistNotFound, "No shoplists found.")
	}

	// Create a map to store unique shoplists
	shoplistMap := make(map[int]*Shoplist)

	// Process results
	for _, r := range results {
		// Add shoplist if not exists
		if _, exists := shoplistMap[r.ShopListID]; !exists {
			shoplistMap[r.ShopListID] = &Shoplist{
				ID:            r.ShopListID,
				Name:          r.ShopListName,
				OwnerID:       r.OwnerID,
				OwnerNickname: r.OwnerNickname,
				Items:         make([]ShoplistItem, 0),
			}
		}

		// Add item if exists
		if r.ItemID != nil {
			item := ShoplistItem{
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
	shoplists := make([]*Shoplist, 0, len(shoplistMap))
	for _, shoplist := range shoplistMap {
		shoplists = append(shoplists, shoplist)
	}

	// Sort shoplists by ID
	sort.Slice(shoplists, func(i, j int) bool {
		return shoplists[i].ID < shoplists[j].ID
	})

	return shoplists, nil
}

func (b *ShoplistBiz) GetShoplistAndItems(ctx context.Context, userID string, shoplistID int) (*Shoplist, *ShoplistError) {
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
		SELECT tbl2.shop_list_id as shop_list_id, shop_list_name, member_id, item_id, item_name, brand_name, extra_info, is_bought, owner_id, owner_nickname 
		FROM (SELECT * FROM shoplist_members WHERE member_id = ?) as tbl1
		LEFT JOIN (
			SELECT shop_list_id, shop_list_name, item_id, item_name, brand_name, extra_info, is_bought, owner_id, nickname as owner_nickname 
			FROM (
				SELECT shop_list_id as shop_list_id, name as shop_list_name, shoplist_items.id as item_id, item_name, brand_name, extra_info, is_bought, owner_id 
				FROM shoplist_items 
				LEFT JOIN shoplists ON shoplist_items.shop_list_id = shoplists.id
			) as shoplistXshoplistItems 
			LEFT JOIN users ON users.id = shoplistXshoplistItems.owner_id
		) as tbl2
		ON tbl2.shop_list_id = tbl1.shop_list_id
		WHERE tbl1.shop_list_id = ?`, userID, shoplistID).Scan(&results).Error

	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to get shoplist items.")
	}

	if len(results) == 0 {
		return nil, NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	// Create shoplist from first row (all rows have same shoplist info)
	shoplist := &Shoplist{
		ID:            results[0].ShopListID,
		Name:          results[0].ShopListName,
		OwnerID:       results[0].OwnerID,
		OwnerNickname: results[0].OwnerNickname,
		Items:         make([]ShoplistItem, 0),
	}

	// Add items to shoplist
	for _, r := range results {
		if r.ItemID != nil {
			shoplist.Items = append(shoplist.Items, ShoplistItem{
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
