package controllers

import (
	"errors"
	"log"
	"mygram-api/dto"
	"mygram-api/helpers"
	"mygram-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserController menyimpan dependensi DB
type UserController struct {
	DB     *gorm.DB
	Logger *log.Logger
}

// NewUserController adalah constructor yang menerima dependensi DB
func NewUserController(db *gorm.DB, appLogger *log.Logger) *UserController {
	return &UserController{
		DB:     db,
		Logger: appLogger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with age, email, password, and username.
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserRegisterRequest true "User registration details"
// @Success 201 {object} dto.UserRegisterResponse
// @Failure 400 {object} dto.BaseResponseError "Invalid request or validation error"
// @Router /users/register [post]
func (u *UserController) Register(c *gin.Context) {
	var req dto.UserRegisterRequest

	// 1. Bind and Validate Request Body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	hashPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	// 2. Create Model Instance with UUID
	user := models.User{
		ID:       uuid.New(), // Modification: Manually set UUID
		Username: req.Username,
		Email:    req.Email,
		Age:      req.Age,
		Password: hashPassword,
	}

	// 3. Save to DB
	if err := u.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to register user",
		})
		return
	}

	// 4. Send Welcome Email (Non-blocking via Goroutine)
	helpers.SendWelcomeEmail(user.Email, user.Username) // Modification: Welcome Email + Goroutine

	// 5. Return Response
	response := dto.UserRegisterResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Age:      user.Age,
	}

	c.JSON(http.StatusCreated, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "User registered successfully",
		Data:    &response,
	}) // Status 201
}

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns a JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserLoginRequest true "User login credentials"
// @Success 200 {object} dto.UserLoginResponse "Successfully logged in"
// @Failure 400 {object} dto.BaseResponseError "Invalid request body or validation error"
// @Failure 401 {object} dto.BaseResponseError "Invalid email or password"
// @Router /users/login [post]
func (u *UserController) Login(c *gin.Context) {
	var req dto.UserLoginRequest // DTO untuk request body
	var user models.User         // Model untuk menampung data user dari DB

	// 1. Binding Request Body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 2. Pengecekan Email di Database
	// Mencari user berdasarkan email. Hanya ambil password untuk pengecekan.
	if err := u.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika email tidak ditemukan
			c.JSON(http.StatusNotFound, dto.BaseResponseError{
				Success: false,
				Message: "Account not found",
			}) // Status 404
			return
		}
		// Error database lainnya
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to login",
		})
		return
	}

	// 3. Pengecekan Password (Menggunakan Bcrypt)
	// req.Password adalah password mentah, user.Password adalah hash dari database
	if isMatch := helpers.CheckPasswordHash(req.Password, user.Password); !isMatch {
		// Jika password tidak cocok
		c.JSON(http.StatusUnauthorized, dto.BaseResponseError{
			Success: false,
			Message: "Invalid email or password",
		}) // Status 401
		return
	}

	// 4. Generate JWT Token
	// Menggunakan helper yang sudah kita buat sebelumnya (helpers/auth.go)
	token, err := helpers.CreateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	// 5. Respon Sukses
	response := dto.UserLoginResponse{
		Token: token, // token: "jwt string"
	}

	c.JSON(http.StatusOK, response) // Status 200
}

// Update godoc
// @Summary Update user's account details
// @Description Update authenticated user's email and username. Requires JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.UserUpdateRequest true "Updated user details"
// @Success 200 {object} dto.BaseResponseSuccessWithData "Successfully updated user account"
// @Failure 400 {object} dto.BaseResponseError "Invalid request or validation error"
// @Failure 401 {object} dto.BaseResponseError "Unauthorized"
// @Failure 500 {object} dto.BaseResponseError "Database error"
// @Router /users [put]
func (u *UserController) Update(c *gin.Context) {
	// Ambil data user dari konteks yang disuntikkan oleh JWT middleware
	userData := c.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr) // ID user yang sedang login

	var req dto.UserUpdateRequest
	var user models.User

	// 1. Binding dan Validasi Request Body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.BaseResponseError{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 2. Buat map data yang akan diupdate
	updatedData := models.User{
		Email:    req.Email,
		Username: req.Username,
	}

	// 3. Update data di database
	// Menggunakan u.DB (dari Dependency Injection)
	// Kita hanya mengupdate Email dan Username
	if err := u.DB.Model(&user).Where("id = ?", userID).Updates(updatedData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to update user account",
		})
		return
	}

	// 4. Ambil data yang sudah diupdate untuk respons
	if err := u.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to retrieve updated data",
		})
		return
	}

	// 5. Respon Sukses (Status 200)
	response := dto.UserUpdateResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Age:       user.Age,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, dto.BaseResponseSuccessWithData{
		Success: true,
		Message: "User account updated successfully",
		Data:    response,
	})
}

// Delete godoc
// @Summary Delete user's account
// @Description Delete the authenticated user's account. Requires JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BaseResponseSuccess "Successfully deleted account"
// @Failure 401 {object} dto.BaseResponseError "Unauthorized"
// @Failure 500 {object} dto.BaseResponseError "Database error"
// @Router /users [delete]
func (u *UserController) Delete(c *gin.Context) {
	// Ambil data user dari konteks
	userData := c.MustGet("userData").(map[string]any)
	userIDStr := userData["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	var user models.User

	// Hapus user berdasarkan ID. Gorm akan menangani penghapusan terkait (CASCADE)
	// jika relasi di model sudah disetel dengan benar.
	if err := u.DB.Where("id = ?", userID).Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.BaseResponseError{
			Success: false,
			Message: "Failed to delete user account",
		})
		return
	}

	// Respon Sukses (Status 200)
	c.JSON(http.StatusOK, dto.BaseResponseSuccess{
		Success: true,
		Message: "Your account has been successfully deleted",
	})
}
