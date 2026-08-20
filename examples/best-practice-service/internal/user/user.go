package user

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusActive Status = "active"
)

type User struct {
	ID        string
	Email     string
	Name      string
	Status    Status
	CreatedAt time.Time
}

func New(cmd RegisterCommand) *User {
	now := time.Now().UTC()

	return &User{
		ID:        fmt.Sprintf("usr_%d", now.UnixNano()),
		Email:     cmd.Email,
		Name:      cmd.Name,
		Status:    StatusActive,
		CreatedAt: now,
	}
}
