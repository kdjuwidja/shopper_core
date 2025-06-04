package bizuser

import (
	"context"

	"github.com/kdjuwidja/aishoppercommon/db"
	"gorm.io/gorm/clause"
	dbmodel "netherealmstudio.com/m/v2/db"
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

func (b *UserBiz) GetUserProfile(ctx context.Context, userID string) (*dbmodel.User, error) {
	user := &dbmodel.User{}
	if err := b.dbPool.GetDB().WithContext(ctx).Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (b *UserBiz) CreateOrUpdateUserProfile(ctx context.Context, userID string, user *dbmodel.User) error {
	if err := b.dbPool.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"postal_code", "nickname"}),
	}).Create(&user).Error; err != nil {
		return err
	}

	return nil
}
