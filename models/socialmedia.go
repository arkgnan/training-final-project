package models

import (
	"time"

	"github.com/google/uuid"
)

type SocialMedia struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	SocialMediaUrl string    `gorm:"not null" json:"social_media_url"`
	UserID         uuid.UUID `json:"user_id"` // Foreign Key of User
	User           *User     `json:"User,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
