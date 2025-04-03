package apiHandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
	"netherrealmstudio.com/aishoppercore/m/util"
)

// getShoplistWithMembers retrieves shoplist data including owner and members
type shoplistData struct {
	ShopListID int
	OwnerID    string
	Members    map[string]struct{ MemberID string }
}

// helper function to get shoplist and members relationship and transform the data into struct for easier use
func getShoplistWithMembers(shoplistID int) (*shoplistData, error) {
	rows, err := db.GetDB().Raw(`SELECT shoplists.id as shop_list_id, shoplists.owner_id as owner_id, shoplist_members.member_id as member_id from shoplists 
		LEFT JOIN shoplist_members ON shoplists.id = shoplist_members.shop_list_id 
		WHERE shoplists.id = ?`, shoplistID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type QueryResult struct {
		ShopListID int    `json:"shop_list_id" gorm:"column:shop_list_id"`
		OwnerID    string `json:"owner_id" gorm:"column:owner_id"`
		MemberID   string `json:"member_id" gorm:"column:member_id"`
	}

	// Transform the response data into the operable shoplist data
	var shopListData shoplistData
	for rows.Next() {
		var queryShoplist QueryResult
		err := rows.Scan(&queryShoplist.ShopListID, &queryShoplist.OwnerID, &queryShoplist.MemberID)
		if err != nil {
			return nil, err
		}

		if shopListData.ShopListID == 0 {
			shopListData.ShopListID = queryShoplist.ShopListID
			shopListData.OwnerID = queryShoplist.OwnerID
			shopListData.Members = make(map[string]struct{ MemberID string })
		}

		if _, exists := shopListData.Members[queryShoplist.MemberID]; !exists {
			shopListData.Members[queryShoplist.MemberID] = struct{ MemberID string }{MemberID: queryShoplist.MemberID}
		}
	}

	return &shopListData, nil
}

// CreateShoplist creates a new shoplist
// @Summary Create a new shoplist
// @Description Creates a new shoplist with the specified name. The authenticated user becomes the owner of the shoplist.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param request body struct{Name string} true "Shoplist details"
// @Success 201 {object} map[string]interface{} "Successfully created shoplist"
// @Failure 400 {object} map[string]string "Name is required"
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to create shoplist"
// @Router /shoplist [put]
func CreateShoplist(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Start a new transaction
	tx := db.GetDB().Begin()

	// Create new shoplist
	shoplist := model.Shoplist{
		OwnerID: userID,
		Name:    req.Name,
	}

	// Save to database
	if err := tx.Create(&shoplist).Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shoplist"})
		return
	}

	// Add the owner as a member of the shoplist
	member := model.ShoplistMember{
		ShopListID: shoplist.ID,
		MemberID:   userID,
	}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add owner as a member"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

// GetAllShoplists retrieves all shoplists for the authenticated user
// @Summary Get all shoplists
// @Description Retrieves all shoplists where the authenticated user is a member, including the owner's information.
// @Tags shoplist
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Successfully retrieved shoplists"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to fetch shoplists"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist [get]
func GetAllShoplists(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get all shoplists where user is a member
	rows, err := db.GetDB().Raw(`
		select shoplists.id as id, shoplists.name as name, users.id as owner_id, users.nickname as owner_nickname from shoplists
		left join users on shoplists.owner_id = users.id
		where shoplists.id in (select shop_list_id from shoplist_members where member_id=?);
	`, userID).Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shoplists"})
		return
	}
	defer rows.Close()

	// Transform the response to only include owner nickname
	type ShoplistResponse struct {
		ID    int    `json:"id" gorm:"column:id"`
		Name  string `json:"name" gorm:"column:name"`
		Owner struct {
			Owner_id string `json:"id" gorm:"column:owner_id"`
			Nickname string `json:"nickname" gorm:"column:owner_nickname"`
		} `json:"owner"`
	}

	response := make([]ShoplistResponse, 0)
	for rows.Next() {
		var r ShoplistResponse
		err := rows.Scan(&r.ID, &r.Name, &r.Owner.Owner_id, &r.Owner.Nickname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
			return
		}
		response = append(response, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"shoplists": response,
	})
}

// GetShoplist retrieves a specific shoplist and its items
// @Summary Get a specific shoplist
// @Description Retrieves a shoplist by its ID along with its items and members
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} ResponseModel
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Failed to fetch shoplist"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist/{id} [get]
func GetShoplist(c *gin.Context) {
	// Get shoplist ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is authorized to access this shoplist
	var member model.ShoplistMember
	err = db.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).First(&member).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	rows, err := db.GetDB().Raw(`SELECT shoplists.id as shop_list_id, shoplists.name as shop_list_name, owner_id, shoplist_items.id as shop_list_item_id, item_name, brand_name, extra_info, is_bought, member_id, member_nickname FROM shoplists
		LEFT JOIN shoplist_items on shoplist_items.shop_list_id = shoplists.id
		LEFT JOIN (SELECT shop_list_id, member_id, nickname as member_nickname from shoplist_members left join users on shoplist_members.member_id = users.id) as tbl1 ON tbl1.shop_list_id = shoplists.id
		where shoplists.id = ?;`, shoplistID).Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shoplist"})
		return
	}
	defer rows.Close()

	// retrieve data from database response
	type PremassageResp struct {
		ID                     int     `json:"id" gorm:"column:shop_list_id"`
		Name                   string  `json:"name" gorm:"column:shop_list_name"`
		OwnerId                string  `json:"owner_id" gorm:"column:owner_id"`
		ShopListItemID         *int    `json:"shop_list_item_id" gorm:"column:shop_list_item_id"`
		ShopListItemName       *string `json:"item_name" gorm:"column:item_name"`
		ShopListItemBrandName  *string `json:"brand_name" gorm:"column:brand_name"`
		ShopListItemExtraInfo  *string `json:"extra_info" gorm:"column:extra_info"`
		ShopListItemIsBought   *bool   `json:"is_bought" gorm:"column:is_bought"`
		ShopListMemberID       string  `json:"member_id" gorm:"column:member_id"`
		ShopListMemberNickname string  `json:"member_nickname" gorm:"column:member_nickname"`
	}

	var premassage_resps []PremassageResp
	for rows.Next() {
		var r PremassageResp
		err := rows.Scan(&r.ID, &r.Name, &r.OwnerId, &r.ShopListItemID, &r.ShopListItemName, &r.ShopListItemBrandName, &r.ShopListItemExtraInfo, &r.ShopListItemIsBought, &r.ShopListMemberID, &r.ShopListMemberNickname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
			return
		}

		premassage_resps = append(premassage_resps, r)
	}

	// massage response data into expected format
	if len(premassage_resps) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	type ResponseModel struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Owner struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"owner"`
		Members []struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"members"`
		Items []struct {
			ID        int    `json:"id"`
			ItemName  string `json:"item_name"`
			BrandName string `json:"brand_name"`
			ExtraInfo string `json:"extra_info"`
			IsBought  bool   `json:"is_bought"`
		} `json:"items"`
	}

	response := ResponseModel{}
	shoplistMembers := make(map[string]struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
	})

	shoplistItems := make(map[int]struct {
		ID        int    `json:"id"`
		ItemName  string `json:"item_name"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
		IsBought  bool   `json:"is_bought"`
	})

	for _, r := range premassage_resps {
		if response.ID == 0 {
			response.ID = r.ID
		}
		if response.Name == "" {
			response.Name = r.Name
		}
		if response.Owner.ID == "" {
			response.Owner.ID = r.OwnerId
		}

		if r.ShopListMemberID != "" {
			if _, exists := shoplistMembers[r.ShopListMemberID]; !exists {
				shoplistMembers[r.ShopListMemberID] = struct {
					ID       string `json:"id"`
					Nickname string `json:"nickname"`
				}{
					ID:       r.ShopListMemberID,
					Nickname: r.ShopListMemberNickname,
				}

				if response.Owner.ID == r.ShopListMemberID {
					response.Owner.Nickname = r.ShopListMemberNickname
				}
			}
		}

		if r.ShopListItemID != nil {
			if _, exists := shoplistItems[*r.ShopListItemID]; !exists {
				shoplistItems[*r.ShopListItemID] = struct {
					ID        int    `json:"id"`
					ItemName  string `json:"item_name"`
					BrandName string `json:"brand_name"`
					ExtraInfo string `json:"extra_info"`
					IsBought  bool   `json:"is_bought"`
				}{
					ID:        *r.ShopListItemID,
					ItemName:  *r.ShopListItemName,
					BrandName: *r.ShopListItemBrandName,
					ExtraInfo: *r.ShopListItemExtraInfo,
					IsBought:  *r.ShopListItemIsBought,
				}
			}
		}
	}

	// convert map to slice
	response.Members = make([]struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
	}, 0)
	for _, member := range shoplistMembers {
		response.Members = append(response.Members, member)
	}

	response.Items = make([]struct {
		ID        int    `json:"id"`
		ItemName  string `json:"item_name"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
		IsBought  bool   `json:"is_bought"`
	}, 0)
	for _, item := range shoplistItems {
		response.Items = append(response.Items, item)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateShoplist updates a shoplist's name
// @Summary Update a shoplist's name
// @Description Updates the name of a specific shoplist by its ID
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param name body string true "Name of the shoplist"
// @Success 200 {object} gin.H
// @Failure 400 {object} map[string]string "Name is required"
// @Failure 403 {object} map[string]string "Only the owner can update this shoplist"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist/{id} [post]
func UpdateShoplist(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if shopListData.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can update this shoplist"})
		return
	}

	// Update shoplist
	if err := db.GetDB().Model(&model.Shoplist{}).Where("id = ?", shoplistID).Update("name", requestBody.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shoplist"})
		return
	}

	// Return updated shoplist
	c.JSON(http.StatusOK, gin.H{})
}

// LeaveShopList allows a user to leave a shoplist
// @Summary Leave a shoplist
// @Description Allows a user to leave a shoplist. If the user is the owner, ownership will be transferred to another member. If the user is the last member, the shoplist will be deleted.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} map[string]interface{} "Successfully left the shoplist"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shoplist/{id}/leave [POST]
func LeaveShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// If no other members, delete the shoplist
	if len(shopListData.Members) == 1 {
		// Use transaction to batch remove member and delete shoplist
		if err := db.GetDB().Transaction(func(tx *gorm.DB) error {
			// First remove the member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
				return err
			}

			// Then delete the shoplist
			if err := tx.Unscoped().Delete(&model.Shoplist{}, shoplistID).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member and delete shoplist"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
		return
	}

	// If user is owner, transfer ownership to another member
	if shopListData.OwnerID == userID {
		// Find another member to transfer ownership to
		var newOwnerID string
		for memberID := range shopListData.Members {
			if memberID != userID {
				newOwnerID = memberID
				break
			}
		}

		// Use transaction to batch transfer ownership and remove member
		if err := db.GetDB().Transaction(func(tx *gorm.DB) error {
			// Transfer ownership
			if err := tx.Model(&model.Shoplist{}).Where("id = ?", shoplistID).Update("owner_id", newOwnerID).Error; err != nil {
				return err
			}

			// Remove member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transfer ownership and remove member"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
		return
	}

	// Remove member
	if err := db.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
}

// RequestShopListShareCode generates a share code for a shoplist
// @Summary Generate share code
// @Description Generates a unique share code for a shoplist that can be used by other users to join. Only the owner can generate share codes. The code expires in 24 hours.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} map[string]interface{} "Successfully generated share code"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Only the owner can generate share codes"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shoplist/{id}/share-code [POST]
func RequestShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if shopListData.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can generate share codes"})
		return
	}

	tx := db.GetDB().Begin()

	// Generate a share code that is unique among all active share codes (6 characters, alphanumeric)
	var shareCode string
	for {
		shareCode = util.GenerateShareCode(6)
		if util.VerifyShareCodeFromDB(tx, shareCode) {
			break
		}
	}
	expiresAt := time.Now().Add(24 * time.Hour) // Share code expires in 24 hours

	// Create or update share code record
	shareCodeRecord := model.ShoplistShareCode{
		ShopListID: shoplistID,
		Code:       shareCode,
		Expiry:     expiresAt,
	}

	// Upsert the share code record
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "shop_list_id"}},
		UpdateAll: true,
	}).Create(&shareCodeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate share code"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_code": shareCode,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// RevokeShopListShareCode revokes the active share code for a shoplist
// @Summary Revoke a share code
// @Description Revokes the active share code for a specific shoplist. Only the owner can revoke share codes.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} gin.H "Successfully revoked share code"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Only the owner can revoke share codes"
// @Failure 404 {object} map[string]string "Not found"
// @Router /shoplist/{id}/share-code/revoke [post]
func RevokeShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Check if user is the owner
	if shopListData.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can revoke share codes"})
		return
	}

	// Find the active share code
	var shareCode model.ShoplistShareCode
	err = db.GetDB().Where("shop_list_id = ? AND expiry > ?", shoplistID, time.Now()).First(&shareCode).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Update the expiry to current time to revoke the code
	if err := db.GetDB().Model(&shareCode).Update("expiry", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke share code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// JoinShopList allows a user to join a shoplist using a share code
// @Summary Join a shoplist using a share code
// @Description Allows a user to join a shoplist by providing a valid share code. The share code must be active and not expired.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param share_code body string true "Share code to join the shoplist"
// @Success 200 {object} gin.H "Successfully joined the shoplist"
// @Failure 400 {object} map[string]string "Share code is required"
// @Failure 400 {object} map[string]string "Invalid share code"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/join [post]
func JoinShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request body
	var requestBody struct {
		ShareCode string `json:"share_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	// Find the active share code
	var shareCode model.ShoplistShareCode
	err := db.GetDB().Where("code = ? AND expiry > ?", requestBody.ShareCode, time.Now()).First(&shareCode).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share code"})
		return
	}

	// Check if user is already a member
	var existingMember model.ShoplistMember
	err = db.GetDB().Where("shop_list_id = ? AND member_id = ?", shareCode.ShopListID, userID).First(&existingMember).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Add user as member
	newMember := model.ShoplistMember{
		ShopListID: shareCode.ShopListID,
		MemberID:   userID,
	}
	if err := db.GetDB().Create(&newMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join shoplist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// AddItemToShopList adds a new item to a shoplist
// @Summary Add a new item to a shoplist
// @Description Adds a new item to a specific shoplist. The user must be a member of the shoplist to add items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
//
//	@Param request body struct {
//	    ItemName  string `json:"item_name" binding:"required"`
//	    BrandName string `json:"brand_name"`
//	    ExtraInfo string `json:"extra_info"`
//	} true "Item details"
//
// @Success 201 {object} map[string]interface{} "Successfully added item"
// @Failure 400 {object} map[string]string "Item name is required"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items [put]
func AddItemToShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		ItemName  string `json:"item_name" binding:"required"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item name is required"})
		return
	}

	// Check if user is a member of the shoplist
	var member model.ShoplistMember
	err = db.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).First(&member).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Create new item
	newItem := model.ShoplistItem{
		ShopListID: shoplistID,
		ItemName:   requestBody.ItemName,
		BrandName:  requestBody.BrandName,
		ExtraInfo:  requestBody.ExtraInfo,
		IsBought:   false,
	}

	// Save to database
	if err := db.GetDB().Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item"})
		return
	}

	// Return the created item
	c.JSON(http.StatusCreated, gin.H{
		"id":         newItem.ID,
		"item_name":  newItem.ItemName,
		"brand_name": newItem.BrandName,
		"extra_info": newItem.ExtraInfo,
		"is_bought":  newItem.IsBought,
	})
}

// RemoveItemFromShopList removes an item from a shoplist
// @Summary Remove an item from a shoplist
// @Description Removes a specific item from a shoplist. The user must be a member of the shoplist to remove items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param itemId path int true "Item ID"
// @Success 200 {object} gin.H "Successfully removed item"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items/{itemId} [delete]
func RemoveItemFromShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Check if item exists and belongs to the shoplist
	var item model.ShoplistItem
	err = db.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check item"})
		return
	}

	// Delete the item
	err = db.GetDB().Unscoped().Delete(&item).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// UpdateShoplistItem updates the bought status of an item
// @Summary Update an item in a shoplist
// @Description Updates the details of a specific item in a shoplist. At least one of the fields (item_name, brand_name, extra_info, is_bought) must be present in the request body. The user must be a member of the shoplist to update items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param itemId path int true "Item ID"
//
//	@Param request body struct {
//	    ItemName  *string `json:"item_name"`
//	    BrandName *string `json:"brand_name"`
//	    ExtraInfo *string `json:"extra_info"`
//	    IsBought  *bool   `json:"is_bought"`
//	} true "Item details"
//
// @Success 200 {object} map[string]interface{} "Successfully updated item"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 400 {object} map[string]string "At least one field must be present in the request body"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items/{itemId} [post]
func UpdateShoplistItem(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		ItemName  *string `json:"item_name"`
		BrandName *string `json:"brand_name"`
		ExtraInfo *string `json:"extra_info"`
		IsBought  *bool   `json:"is_bought"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if request body is empty
	if requestBody.ItemName == nil && requestBody.BrandName == nil &&
		requestBody.ExtraInfo == nil && requestBody.IsBought == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must include at least one of item_name, brand_name, extra_info or Is_bought."})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Check if item exists and belongs to the shoplist
	var item model.ShoplistItem
	err = db.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check item"})
		return
	}

	// Update only the fields that are provided in the request
	updates := make(map[string]interface{})
	if requestBody.ItemName != nil {
		updates["item_name"] = *requestBody.ItemName
	}
	if requestBody.BrandName != nil {
		updates["brand_name"] = *requestBody.BrandName
	}
	if requestBody.ExtraInfo != nil {
		updates["extra_info"] = *requestBody.ExtraInfo
	}
	if requestBody.IsBought != nil {
		updates["is_bought"] = *requestBody.IsBought
	}

	// Update the item
	err = db.GetDB().Model(&item).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	// Return the updated item
	c.JSON(http.StatusOK, gin.H{
		"id":         item.ID,
		"item_name":  item.ItemName,
		"brand_name": item.BrandName,
		"extra_info": item.ExtraInfo,
		"is_bought":  item.IsBought,
	})
}
