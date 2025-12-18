package dto

import "time"

// CommentCreateRequest represents the request body for POST /comments
type CommentCreateRequest struct {
	PhotoID string `json:"photo_id" binding:"required" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Message string `json:"message" binding:"required" example:"Nice shot!"`
}

// CommentReplyRequest represents the request body for POST /comments/reply/:parentCommentID
// CommentUpdateRequest represents the request body for PUT /comments/{commentID}
// ParentCommentID is provided as a URL parameter; the body contains the reply message.
type CommentReplyRequest struct {
	Message string `json:"message" binding:"required" example:"Thanks for sharing!"`
}

// CommentResponse represents the response body for comment resources
type CommentResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	PhotoID         string    `json:"photo_id"`
	Message         string    `json:"message"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	RepliesCount    int       `json:"replies_count"` // Untuk efisiensi
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
