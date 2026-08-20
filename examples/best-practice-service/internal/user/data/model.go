package data

import "time"

type UserModel struct {
	ID string `gorm:"primaryKey;size:32"`

	Email string `gorm:"size:255;uniqueIndex;not null"`
	Name  string `gorm:"size:100;not null"`

	Status string `gorm:"size:32;index;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
