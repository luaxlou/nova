package data

import (
	"context"
	"strings"
	"time"
)

type InsertParams struct {
	ID     string
	Email  string
	Name   string
	Status string
}

func Insert(ctx context.Context, p InsertParams) error {
	if err := ctx.Err(); err != nil {
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

	storeMu.Lock()
	defer storeMu.Unlock()

	usersByEmail[strings.ToLower(p.Email)] = model
	return nil
}
