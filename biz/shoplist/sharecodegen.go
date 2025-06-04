package bizshoplist

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
	"netherealmstudio.com/m/v2/db"
)

// generateShareCode creates a random alphanumeric code of specified length
func GenerateShareCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// VerifyShareCodeFromDB verifies if a share code exists in the database and is not expired
// Returns true if the share code exists and is not expired, false otherwise
// gormDB Context already established before calling this function
func VerifyShareCodeFromDB(gormDB *gorm.DB, code string) bool {
	var existingCode db.ShoplistShareCode
	err := gormDB.Where("code = ? and expiry > ?", code, time.Now()).First(&existingCode).Error
	return err == gorm.ErrRecordNotFound
}
