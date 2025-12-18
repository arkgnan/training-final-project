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

// CommentController menyimpan dependensi DB
type CommentController struct {
	DB     *gorm.DB
	Logger *log.Logger
}

// NewCommentController adalah constructor yang menerima dependensi DB
func NewCommentController(db *gorm.DB, appLogger *log.Logger) *CommentController {
	return &CommentController{
		DB:     db,
		Logger: appLogger,
	}
}

// Create godoc
// @Summary Create a new comment
// @Description Create a new comment for the authenticated user
// @Tags comments
// @Accept json
// @Produce json
// @Param comment body dto.CommentCreateRequest true "Comment create payload"
// @Success 201 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments [post]
func (cc *CommentController) Create(c *gin.Context) {
	// Ambil user dari context (dimasukkan oleh middleware Authentication)
	userData := c.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.CommentCreateRequest
	// Binding dan validasi
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Parse PhotoID
	photoID, err := uuid.Parse(req.PhotoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid photo ID",
		})
		return
	}

	comment := models.Comment{
		ID:      uuid.New(),
		UserID:  userID,
		PhotoID: photoID,
		Message: req.Message,
	}

	if err := cc.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to create comment",
		})
		return
	}

	resp := dto.CommentResponse{
		ID:        comment.ID.String(),
		UserID:    comment.UserID.String(),
		PhotoID:   comment.PhotoID.String(),
		Message:   comment.Message,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	c.JSON(http.StatusCreated, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Comment created successfully",
		Data:    resp,
	})
}

// GetAll godoc
// @Summary Get all comments
// @Description Retrieve all comments
// @Tags comments
// @Produce json
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments [get]
func (cc *CommentController) GetAll(c *gin.Context) {
	var comments []models.Comment

	// Preload User and Photo to include related data if desired
	if err := cc.DB.
		Where("parent_comment_id IS NULL").
		Preload("Replies"). // preload replies untuk semua comments hasil query (batch)
		Preload("User").
		Preload("Photo").
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve comments",
		})
		return
	}

	var respList []dto.CommentResponse
	for _, cm := range comments {
		respList = append(respList, dto.CommentResponse{
			ID:           cm.ID.String(),
			UserID:       cm.UserID.String(),
			PhotoID:      cm.PhotoID.String(),
			Message:      cm.Message,
			RepliesCount: len(cm.Replies),
			CreatedAt:    cm.CreatedAt,
			UpdatedAt:    cm.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Comments retrieved successfully",
		Data:    respList,
	})
}

// Update godoc
// @Summary Update a comment
// @Description Update a comment by id. Requires authorization middleware to ensure ownership.
// @Tags comments
// @Accept json
// @Produce json
// @Param comment body dto.CommentReplyRequest true "Comment update payload"
// @Param commentID path string true "Comment ID"
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments/{commentID} [put]
func (cc *CommentController) Update(c *gin.Context) {
	commentIDStr := c.Param("commentID")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid comment ID",
		})
		return
	}

	var req dto.CommentReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var comment models.Comment
	if err := cc.DB.First(&comment, "id = ?", commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Comment not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve comment",
		})
		return
	}

	updatedData := models.Comment{
		Message: req.Message,
	}

	if err := cc.DB.Model(&comment).Updates(updatedData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to update comment",
		})
		return
	}

	// Ambil kembali data setelah update
	if err := cc.DB.Preload("User").Preload("Photo").First(&comment, "id = ?", commentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve updated comment",
		})
		return
	}

	resp := dto.CommentResponse{
		ID:        comment.ID.String(),
		UserID:    comment.UserID.String(),
		PhotoID:   comment.PhotoID.String(),
		Message:   comment.Message,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Comment updated successfully",
		Data:    resp,
	})
}

// Delete godoc
// @Summary Delete a comment
// @Description Delete a comment by id. Requires authorization middleware to ensure ownership.
// @Tags comments
// @Param commentID path string true "Comment ID"
// @Success 200 {object} dto.BaseResponseSuccess
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments/{commentID} [delete]
func (cc *CommentController) Delete(c *gin.Context) {
	commentIDStr := c.Param("commentID")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid comment ID",
		})
		return
	}

	// Pastikan record ada sebelum mencoba delete (memberikan pesan 404 jika tidak ada)
	var comment models.Comment
	if err := cc.DB.First(&comment, "id = ?", commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Comment not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve comment",
		})
		return
	}

	if err := cc.DB.Where("id = ?", commentID).Delete(&models.Comment{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to delete comment",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccess{
		Success: true,
		Message: "Comment deleted successfully",
	})
}

// CreateReply godoc
// @Summary Create a reply to a comment
// @Description Create a reply comment for an existing parent comment
// @Tags comments
// @Accept json
// @Produce json
// @Param parentCommentID path string true "Parent Comment ID"
// @Param comment body dto.CommentReplyRequest true "Reply payload (message only)"
// @Success 201 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments/reply/{parentCommentID} [post]
func (cc *CommentController) CreateReply(ctx *gin.Context) {
	// 1. Ambil Parent ID dari URL
	parentIDStr := ctx.Param("parentCommentID")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid parent comment ID",
		})
		return
	}

	// 2. Pastikan komentar induk ada
	var parent models.Comment
	if err := cc.DB.First(&parent, "id = ?", parentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Parent comment not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve parent comment",
		})
		return
	}

	// 3. Ambil user dari context
	userData := ctx.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	// 4. Bind body (only message expected)
	var req dto.CommentReplyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 5. Create reply comment. PhotoID is taken from parent to keep consistency.
	reply := models.Comment{
		ID:              uuid.New(),
		UserID:          userID,
		PhotoID:         parent.PhotoID,
		Message:         req.Message,
		ParentCommentID: &parentID,
	}

	if err := cc.DB.Create(&reply).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to create reply",
		})
		return
	}

	// Prepare parent pointer string for response
	var parentIDStrOut *string
	if reply.ParentCommentID != nil {
		s := (*reply.ParentCommentID).String()
		parentIDStrOut = &s
	}

	resp := dto.CommentResponse{
		ID:              reply.ID.String(),
		UserID:          reply.UserID.String(),
		PhotoID:         reply.PhotoID.String(),
		Message:         reply.Message,
		ParentCommentID: parentIDStrOut,
		RepliesCount:    0,
		CreatedAt:       reply.CreatedAt,
		UpdatedAt:       reply.UpdatedAt,
	}

	ctx.JSON(http.StatusCreated, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Reply created successfully",
		Data:    resp,
	})
}

// GetReplies mengambil semua balasan dari komentar induk
// @Summary Get replies for a parent comment
// @Description Get replies for a given parent comment
// @Tags comments
// @Produce json
// @Param parentCommentID path string true "Parent Comment ID"
// @Success 200 {object} dto.BaseResponseSuccessWithData
// @Failure 400 {object} dto.BaseResponseError
// @Failure 404 {object} dto.BaseResponseError
// @Failure 500 {object} dto.BaseResponseError
// @Security BearerAuth
// @Router /comments/{parentCommentID}/replies [get]
func (cc *CommentController) GetReplies(ctx *gin.Context) {
	parentIDStr := ctx.Param("parentCommentID")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: "Invalid parent comment ID",
		})
		return
	}

	// Pastikan parent ada
	var parent models.Comment
	if err := cc.DB.First(&parent, "id = ?", parentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Parent comment not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve parent comment",
		})
		return
	}

	// Ambil semua balasan untuk parent tersebut
	var replies []models.Comment
	if err := cc.DB.Where("parent_comment_id = ?", parentID).Preload("User").Preload("Replies").Find(&replies).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve replies",
		})
		return
	}

	var resp []dto.CommentResponse
	for _, r := range replies {
		var parentIDOut *string
		if r.ParentCommentID != nil {
			p := (*r.ParentCommentID).String()
			parentIDOut = &p
		}
		resp = append(resp, dto.CommentResponse{
			ID:              r.ID.String(),
			UserID:          r.UserID.String(),
			PhotoID:         r.PhotoID.String(),
			Message:         r.Message,
			ParentCommentID: parentIDOut,
			RepliesCount:    len(r.Replies),
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "Replies retrieved successfully",
		Data:    resp,
	})
}
