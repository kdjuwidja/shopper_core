package bizuser

import (
	"gorm.io/gorm/clause"
	"netherrealmstudio.com/aishoppercore/m/db"
)

type UserBiz struct {
	dbPool db.MySQLConnectionPool
}

// Dependency Injection for UserBiz
func InitializeUserBiz(dbPool db.MySQLConnectionPool) *UserBiz {
	return &UserBiz{
		dbPool: dbPool,
	}
}

func (b *UserBiz) GetUserProfile(userID string) (*db.User, error) {
	user := &db.User{}
	if err := b.dbPool.GetDB().Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (b *UserBiz) CreateOrUpdateUserProfile(userID string, user *db.User) error {
	if err := b.dbPool.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"postal_code", "nickname"}),
	}).Create(&user).Error; err != nil {
		return err
	}

	return nil
}
