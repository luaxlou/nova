package data

import "time"

type UserModel struct {
	ID string `gorm:"column:id;primaryKey;size:32"`

	Email string `gorm:"column:email;size:255;uniqueIndex;not null"`
	Name  string `gorm:"column:name;size:100;not null"`

	Status string `gorm:"column:status;size:32;index;not null"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (UserModel) TableName() string {
	return "users"
}
