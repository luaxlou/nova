package data

import (
	"context"

	"github.com/luaxlou/nova/starter/novagorm"
)

func EmailExists(ctx context.Context, email string) (bool, error) {
	db, err := novagorm.DB()
	if err != nil {
		return false, err
	}

	var count int64
	err = db.WithContext(ctx).
		Model(&UserModel{}).
		Where("email = ?", email).
		Count(&count).
		Error

	return count > 0, err
}
