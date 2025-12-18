package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID          uuid.UUID  `json:"user_id"`  // Foreign Key of User
	PhotoID         uuid.UUID  `json:"photo_id"` // Foreign Key of Photo
	Message         string     `gorm:"not null" json:"message"`
	ParentCommentID *uuid.UUID `gorm:"type:uuid;default:null" json:"parent_comment_id,omitempty"`  // FK ke Comment.ID
	ParentComment   *Comment   `gorm:"foreignkey:ParentCommentID" json:"parent_comment,omitempty"` // Relasi ke komentar induk
	Replies         []Comment  `gorm:"foreignkey:ParentCommentID" json:"replies,omitempty"`        // Relasi balasan
	User            *User      `json:"User,omitempty"`
	Photo           *Photo     `json:"Photo,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
