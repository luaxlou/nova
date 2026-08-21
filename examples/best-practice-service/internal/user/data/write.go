package data

import (
	"context"
	"time"

	"github.com/luaxlou/nova/orm/novagorm"
)

type InsertParams struct {
	ID     string
	Email  string
	Name   string
	Status string
}

func Insert(ctx context.Context, p InsertParams) error {
	db, err := novagorm.DB()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	model := UserModel{
		ID:        p.ID,
		Email:     p.Email,
		Name:      p.Name,
		Status:    p.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return db.WithContext(ctx).Create(&model).Error
}
