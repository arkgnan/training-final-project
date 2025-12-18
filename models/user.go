package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"` // Modification: Use UUID
	Username     string        `gorm:"not null;unique" json:"username"`
	Email        string        `gorm:"not null;unique" json:"email"`
	Password     string        `gorm:"not null" json:"password"`
	Age          int           `gorm:"not null" json:"age"`
	Photos       []Photo       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"photos"`
	Comments     []Comment     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"comments"`
	SocialMedias []SocialMedia `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"social_medias"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
