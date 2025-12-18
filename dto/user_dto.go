package dto

import "time"

// UserRegisterRequest represents the request body for POST /users/register
type UserRegisterRequest struct {
	Age      int    `json:"age" binding:"required,numeric,gte=9,lte=100"`
	Email    string `json:"email" binding:"required,email,uniqueEmail" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Username string `json:"username" binding:"required,uniqueUsername" example:"john_doe"`
}

// UserRegisterResponse represents the successful response for POST /users/register
type UserRegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

// UserUpdateRequest merepresentasikan request body untuk PUT /users
type UserUpdateRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
}

// UserUpdateResponse merepresentasikan response body sukses untuk PUT /users
type UserUpdateResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	UpdatedAt time.Time `json:"updated_at"`
}
