package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         string `json:"id" gorm:"type:varchar(32);primaryKey"`
	Nickname   string `json:"nickname" gorm:"type:varchar(100);not null"`
	PostalCode string `json:"postal_code" gorm:"type:varchar(6);not null"`
}

type Shoplist struct {
	gorm.Model
	ID      int    `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:true;not null;AUTO_INCREMENT:10000"`
	OwnerID string `json:"-" gorm:"not null"`
	Owner   User   `json:"owner" gorm:"foreignKey:OwnerID;reference:ID"`
	Name    string `json:"name" gorm:"type:varchar(100);not null"`
}

type ShoplistShareCode struct {
	gorm.Model
	ID         int       `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:true;not null;AUTO_INCREMENT:10000"`
	ShopListID int       `json:"-" gorm:"not null;unique"`
	ShopList   Shoplist  `json:"shoplist" gorm:"foreignKey:ShopListID;reference:ID"`
	Code       string    `json:"code" gorm:"type:varchar(6);not null"`
	Expiry     time.Time `json:"expiry" gorm:"type:timestamp;"`
}

type ShoplistMember struct {
	gorm.Model
	ID         int      `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:true;not null;AUTO_INCREMENT:10000"`
	ShopListID int      `json:"-" gorm:"not null;uniqueIndex:idx_shoplist_member"`
	ShopList   Shoplist `json:"shoplist" gorm:"foreignKey:ShopListID;reference:ID"`
	MemberID   string   `json:"-" gorm:"not null;uniqueIndex:idx_shoplist_member"`
	Member     User     `json:"member" gorm:"foreignKey:MemberID;reference:ID"`
}

type ShoplistItem struct {
	gorm.Model
	ID         int      `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:true;not null;AUTO_INCREMENT:10000"`
	ShopListID int      `json:"-" gorm:"not null"`
	ShopList   Shoplist `json:"shoplist" gorm:"foreignKey:ShopListID;reference:ID"`
	ItemName   string   `json:"item_name" gorm:"type:varchar(100);not null"`
	BrandName  string   `json:"brand_name" gorm:"type:varchar(100);not null"`
	ExtraInfo  string   `json:"extra_info" gorm:"type:varchar(100);"`
	IsBought   bool     `json:"is_bought" gorm:"type:tinyint(1);not null;default:0"`
}
