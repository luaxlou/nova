package user

import (
	"context"
	"strings"

	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/data"
)

type RegisterCommand struct {
	Email string
	Name  string
}

func Register(ctx context.Context, cmd RegisterCommand) (*User, error) {
	cmd.Email = strings.TrimSpace(cmd.Email)
	cmd.Name = strings.TrimSpace(cmd.Name)

	if err := validateRegister(cmd); err != nil {
		return nil, err
	}

	exists, err := data.EmailExists(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailExists
	}

	u := New(cmd)

	if err := data.Insert(ctx, data.InsertParams{
		ID:     u.ID,
		Email:  u.Email,
		Name:   u.Name,
		Status: string(u.Status),
	}); err != nil {
		return nil, err
	}

	return u, nil
}

func validateRegister(cmd RegisterCommand) error {
	if cmd.Email == "" {
		return ErrInvalidEmail
	}
	if cmd.Name == "" {
		return ErrInvalidName
	}
	return nil
}
