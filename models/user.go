package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"` // Modification: Use UUID
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

// BeforeCreate hook will set a UUID in application code if it's not already set.
// This avoids DB-specific defaults like uuid_generate_v4() and works on SQLite.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
