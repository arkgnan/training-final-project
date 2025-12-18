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

// PhotoController menyimpan dependensi DB
type PhotoController struct {
	DB     *gorm.DB
	Logger *log.Logger
}

// NewPhotoController adalah constructor yang menerima dependensi DB
func NewPhotoController(db *gorm.DB, appLogger *log.Logger) *PhotoController {
	return &PhotoController{
		DB:     db,
		Logger: appLogger,
	}
}

// Using Photo DTOs from dto/photoDto.go
// PhotoCreateRequest, PhotoUpdateRequest and PhotoResponse are defined in the dto package

// Create godoc
// @Summary Create a new photo
// @Description Create a new photo for authenticated user
// @Tags photos
// @Accept json
// @Produce json
// @Param photo body dto.PhotoUpsertRequest true "Photo create payload"
// @Success 201 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /photos [post]
func (p *PhotoController) Create(c *gin.Context) {
	// Ambil user dari context (dimasukkan oleh middleware Authentication)
	userData := c.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.PhotoUpsertRequest

	// Binding dan validasi
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	photo := models.Photo{
		ID:       uuid.New(),
		Title:    req.Title,
		Caption:  req.Caption,
		PhotoUrl: req.PhotoUrl,
		UserID:   userID,
	}

	if err := p.DB.Create(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to create photo",
		})
		return
	}

	resp := dto.PhotoResponse{
		ID:        photo.ID.String(),
		Title:     photo.Title,
		Caption:   photo.Caption,
		PhotoUrl:  photo.PhotoUrl,
		UserID:    photo.UserID.String(),
		CreatedAt: photo.CreatedAt,
		UpdatedAt: photo.UpdatedAt,
	}

	c.JSON(http.StatusCreated, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Photo created successfully",
		Data:    resp,
	})
}

// GetAll godoc
// @Summary Get all photos
// @Description Retrieve all photos (with owner info)
// @Tags photos
// @Produce json
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /photos [get]
func (p *PhotoController) GetAll(c *gin.Context) {
	var photos []models.Photo

	// Preload User and Comments to include related data
	if err := p.DB.Preload("User").Preload("Comments").Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve photos",
		})
		return
	}

	var respList []dto.PhotoResponse
	for _, ph := range photos {
		respList = append(respList, dto.PhotoResponse{
			ID:        ph.ID.String(),
			Title:     ph.Title,
			Caption:   ph.Caption,
			PhotoUrl:  ph.PhotoUrl,
			UserID:    ph.UserID.String(),
			CreatedAt: ph.CreatedAt,
			UpdatedAt: ph.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Photos retrieved successfully",
		Data:    respList,
	})
}

// Update godoc
// @Summary Update a photo
// @Description Update a photo by id. Requires authorization middleware to ensure ownership.
// @Tags photos
// @Accept json
// @Produce json
// @Param photo body dto.PhotoUpsertRequest true "Photo update payload"
// @Param photoID path string true "Photo ID"
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /photos/{photoID} [put]
func (p *PhotoController) Update(c *gin.Context) {
	photoIDStr := c.Param("photoID")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid photo ID",
		})
		return
	}

	var req dto.PhotoUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var photo models.Photo
	if err := p.DB.First(&photo, "id = ?", photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Photo not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve photo",
		})
		return
	}

	updatedData := models.Photo{
		Title:    req.Title,
		Caption:  req.Caption,
		PhotoUrl: req.PhotoUrl,
	}

	if err := p.DB.Model(&photo).Updates(updatedData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to update photo",
		})
		return
	}

	// Ambil kembali data setelah update
	if err := p.DB.Preload("User").First(&photo, "id = ?", photoID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve updated photo",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Photo updated successfully",
		Data: dto.PhotoResponse{
			ID:        photo.ID.String(),
			Title:     photo.Title,
			Caption:   photo.Caption,
			PhotoUrl:  photo.PhotoUrl,
			UserID:    photo.UserID.String(),
			CreatedAt: photo.CreatedAt,
			UpdatedAt: photo.UpdatedAt,
		},
	})
}

// Delete godoc
// @Summary Delete a photo
// @Description Delete a photo by id. Requires authorization middleware to ensure ownership.
// @Tags photos
// @Param photoID path string true "Photo ID"
// @Success 200 {object} dto.BaseResponseSuccess
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /photos/{photoID} [delete]
func (p *PhotoController) Delete(c *gin.Context) {
	photoIDStr := c.Param("photoID")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid photo ID",
		})
		return
	}

	// Pastikan record ada sebelum mencoba delete (memberikan pesan 404 jika tidak ada)
	var photo models.Photo
	if err := p.DB.First(&photo, "id = ?", photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Photo not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve photo",
		})
		return
	}

	if err := p.DB.Where("id = ?", photoID).Delete(&models.Photo{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to delete photo",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccess{
		Success: true,
		Message: "Photo deleted successfully",
	})
}
