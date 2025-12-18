package dto

import "time"

// SocialMediaCreateRequest represents the request body for POST /socialmedias
type SocialMediaCreateRequest struct {
	Name           string `json:"name" binding:"required" example:"Instagram"`
	SocialMediaUrl string `json:"social_media_url" binding:"required,url" example:"https://instagram.com/yourhandle"`
}

// SocialMediaUpdateRequest represents the request body for PUT /socialmedias/{socialMediaID}
type SocialMediaUpdateRequest struct {
	Name           string `json:"name" binding:"required" example:"Instagram (edited)"`
	SocialMediaUrl string `json:"social_media_url" binding:"required,url" example:"https://instagram.com/yourhandle_updated"`
}

// SocialMediaResponse represents the response body for social media resources
type SocialMediaResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	SocialMediaUrl string    `json:"social_media_url"`
	UserID         string    `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
