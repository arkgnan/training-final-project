package dto

import "time"

// PhotoUpsertRequest represents the request body for POST /photos
type PhotoUpsertRequest struct {
	Title    string `json:"title" binding:"required" example:"Sunset over the beach"`
	Caption  string `json:"caption" example:"A beautiful sunset captured at the shore"`
	PhotoUrl string `json:"photo_url" binding:"required,url" example:"https://example.com/photos/1.jpg"`
}

// PhotoResponse represents the response body for photo resources
type PhotoResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Caption   string    `json:"caption"`
	PhotoUrl  string    `json:"photo_url"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
