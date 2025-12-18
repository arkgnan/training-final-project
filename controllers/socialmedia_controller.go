package controllers

import (
	"errors"
	"log"
	"mygram-api/dto"
	"mygram-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialMediaController menyimpan dependensi DB
type SocialMediaController struct {
	DB     *gorm.DB
	Logger *log.Logger
}

// NewSocialMediaController adalah constructor yang menerima dependensi DB
func NewSocialMediaController(db *gorm.DB, appLogger *log.Logger) *SocialMediaController {
	return &SocialMediaController{
		DB:     db,
		Logger: appLogger,
	}
}

// Create godoc
// @Summary Create a new social media entry
// @Description Create a new social media record for the authenticated user
// @Tags socialmedias
// @Accept json
// @Produce json
// @Param socialMedia body dto.SocialMediaCreateRequest true "Social media create payload"
// @Success 201 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /socialmedias [post]
func (smc *SocialMediaController) Create(c *gin.Context) {
	// Ambil user dari context (dimasukkan oleh middleware Authentication)
	userData := c.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.SocialMediaCreateRequest
	// Binding dan validasi
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	social := models.SocialMedia{
		ID:             uuid.New(),
		Name:           req.Name,
		SocialMediaUrl: req.SocialMediaUrl,
		UserID:         userID,
	}

	if err := smc.DB.Create(&social).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to create social media",
		})
		return
	}

	resp := dto.SocialMediaResponse{
		ID:             social.ID.String(),
		Name:           social.Name,
		SocialMediaUrl: social.SocialMediaUrl,
		UserID:         social.UserID.String(),
		CreatedAt:      social.CreatedAt,
		UpdatedAt:      social.UpdatedAt,
	}

	c.JSON(http.StatusCreated, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Social media created successfully",
		Data:    resp,
	})
}

// GetAll godoc
// @Summary Get all social media entries
// @Description Retrieve all social media entries
// @Tags socialmedias
// @Produce json
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /socialmedias [get]
func (smc *SocialMediaController) GetAll(c *gin.Context) {
	var socials []models.SocialMedia
	if err := smc.DB.Preload("User").Find(&socials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve social medias",
		})
		return
	}

	var respList []dto.SocialMediaResponse
	for _, s := range socials {
		respList = append(respList, dto.SocialMediaResponse{
			ID:             s.ID.String(),
			Name:           s.Name,
			SocialMediaUrl: s.SocialMediaUrl,
			UserID:         s.UserID.String(),
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Social medias retrieved successfully",
		Data:    respList,
	})
}

// Update godoc
// @Summary Update a social media entry
// @Description Update a social media entry by id. Authorization middleware should ensure ownership.
// @Tags socialmedias
// @Accept json
// @Produce json
// @Param socialMedia body dto.SocialMediaUpdateRequest true "Social media update payload"
// @Param socialMediaID path string true "Social media ID"
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /socialmedias/{socialMediaID} [put]
func (smc *SocialMediaController) Update(c *gin.Context) {
	socialIDStr := c.Param("socialMediaID")
	socialID, err := uuid.Parse(socialIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid social media ID",
		})
		return
	}

	var req dto.SocialMediaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var social models.SocialMedia
	if err := smc.DB.First(&social, "id = ?", socialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Social media not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve social media",
		})
		return
	}

	updatedData := models.SocialMedia{
		Name:           req.Name,
		SocialMediaUrl: req.SocialMediaUrl,
	}

	if err := smc.DB.Model(&social).Updates(updatedData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to update social media",
		})
		return
	}

	// Ambil kembali data setelah update
	if err := smc.DB.Preload("User").First(&social, "id = ?", socialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve updated social media",
		})
		return
	}

	resp := dto.SocialMediaResponse{
		ID:             social.ID.String(),
		Name:           social.Name,
		SocialMediaUrl: social.SocialMediaUrl,
		UserID:         social.UserID.String(),
		CreatedAt:      social.CreatedAt,
		UpdatedAt:      social.UpdatedAt,
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Social media updated successfully",
		Data:    resp,
	})
}

// Delete godoc
// @Summary Delete a social media entry
// @Description Delete a social media entry by id. Authorization middleware should ensure ownership.
// @Tags socialmedias
// @Param socialMediaID path string true "Social media ID"
// @Success 200 {object} dto.BaseResponseSuccess
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /socialmedias/{socialMediaID} [delete]
func (smc *SocialMediaController) Delete(c *gin.Context) {
	socialIDStr := c.Param("socialMediaID")
	socialID, err := uuid.Parse(socialIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid social media ID",
		})
		return
	}

	var social models.SocialMedia
	if err := smc.DB.First(&social, "id = ?", socialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Social media not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve social media",
		})
		return
	}

	if err := smc.DB.Where("id = ?", socialID).Delete(&models.SocialMedia{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to delete social media",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccess{
		Success: true,
		Message: "Social media deleted successfully",
	})
}
