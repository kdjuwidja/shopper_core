package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         string `json:"id" gorm:"type:varchar(32);primaryKey"`
	PostalCode string `json:"postal_code" gorm:"type:varchar(6);not null"`
}

type Shoplist struct {
	gorm.Model
	ID              int       `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:10000;not null"`
	OwnerID         string    `json:"-" gorm:"not null"`
	Owner           User      `json:"owner" gorm:"foreignKey:OwnerID;reference:ID"`
	Name            string    `json:"name" gorm:"type:varchar(100);not null"`
	ShareCode       string    `json:"share_code" gorm:"type:varchar(6);"`
	ShareCodeExpiry time.Time `json:"share_code_expiry" gorm:"type:timestamp;"`
}

type ShopListMembers struct {
	gorm.Model
	ID         int      `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:10000;not null"`
	ShopListID int      `json:"-" gorm:"not null"`
	ShopList   Shoplist `json:"shoplist" gorm:"foreignKey:ShopListID;reference:ID"`
	MemberID   string   `json:"-" gorm:"not null"`
	Member     User     `json:"member" gorm:"foreignKey:MemberID;reference:ID"`
}

type ShoplistItems struct {
	gorm.Model
	ID         int      `json:"id" gorm:"type:int unsigned;primaryKey;autoIncrement:10000;not null"`
	ShopListID int      `json:"-" gorm:"not null"`
	ShopList   Shoplist `json:"shoplist" gorm:"foreignKey:ShopListID;reference:ID"`
	ItemName   string   `json:"item_name" gorm:"type:varchar(100);not null"`
	BrandName  string   `json:"brand_name" gorm:"type:varchar(100);not null"`
	ExtraInfo  string   `json:"extra_info" gorm:"type:varchar(100);"`
	IsBought   bool     `json:"is_bought" gorm:"type:tinyint(1);not null;default:0"`
}
