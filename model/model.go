package model

import "time"

type User struct {
	ID         string    `json:"id" gorm:"type:varchar(255);primaryKey"`
	PostalCode string    `json:"postal_code"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
