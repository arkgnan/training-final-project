package models

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title     string    `gorm:"not null" json:"title"`
	Caption   string    `json:"caption"`
	PhotoUrl  string    `gorm:"not null" json:"photo_url"`
	UserID    uuid.UUID `json:"user_id"` // Foreign Key of User
	User      *User     `json:"User,omitempty"`
	Comments  []Comment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"comments"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
