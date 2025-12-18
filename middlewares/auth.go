package middlewares

import (
	"maps"
	"mygram-api/database"
	"mygram-api/dto"
	"mygram-api/helpers"
	"mygram-api/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Authentication is a middleware to verify the JWT token
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.BaseResponseError{
				Success: false,
				Message: "Bearer token is required",
			})
			return
		}

		// Check for "Bearer " prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.BaseResponseError{
				Success: false,
				Message: "Token must be a Bearer token",
			})
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]

		// Call helper to verify and parse token
		claims, err := helpers.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.BaseResponseError{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Convert claims (jwt.MapClaims) into a plain map[string]any to avoid named-type
		userData := map[string]any{}
		maps.Copy(userData, claims)

		// Store user data (ID) in context for controllers/authorization
		c.Set("userData", userData)
		c.Next()
	}
}

// Authorization checks if the authenticated user owns the resource
func Authorization(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		userData := c.MustGet("userData").(map[string]any)
		userID := uuid.MustParse(userData["id"].(string))

		resourceIDStr := c.Param(resourceType + "ID") // e.g., "photoID"
		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, dto.BaseResponseError{
				Success: false,
				Message: "Invalid ID format",
			})
			return
		}

		var ownedID uuid.UUID

		// Use a switch to determine the model and query
		switch resourceType {
		case "photo":
			var photo models.Photo
			if err := db.Select("user_id").First(&photo, "id = ?", resourceID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, dto.BaseResponseError{
					Success: false,
					Message: "Photo not found",
				})
				return
			}
			ownedID = photo.UserID
		case "comment":
			var comment models.Comment
			if err := db.Select("user_id").First(&comment, "id = ?", resourceID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, dto.BaseResponseError{
					Success: false,
					Message: "Comment not found",
				})
				return
			}
			ownedID = comment.UserID
		case "socialmedia":
			var sm models.SocialMedia
			if err := db.Select("user_id").First(&sm, "id = ?", resourceID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, dto.BaseResponseError{
					Success: false,
					Message: "Social Media not found",
				})
				return
			}
			ownedID = sm.UserID
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.BaseResponseError{
				Success: false,
				Message: "Invalid resource type",
			})
			return
		}

		// Authorization check
		if ownedID != userID {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.BaseResponseError{
				Success: false,
				Message: "You are not authorized to modify this " + resourceType,
			})
			return
		}

		c.Next()
	}
}
