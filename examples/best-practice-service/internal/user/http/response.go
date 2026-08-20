package userhttp

import (
	"time"

	"github.com/luaxlou/nova/examples/best-practice-service/internal/user"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
	}
}
