package helpers

import (
	"mygram-api/database"
	"mygram-api/models"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func RegisterCustomValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		// --- 1. Custom Validator untuk Email ---
		v.RegisterValidation("uniqueEmail", func(fl validator.FieldLevel) bool {
			email := fl.Field().String()
			var user models.User

			// Query DB menggunakan dbClient global
			db := database.GetDB()
			err := db.Where("email = ?", email).First(&user).Error

			// Return TRUE jika error adalah ErrRecordNotFound (artinya unik)
			return gorm.ErrRecordNotFound == err
		})

		// --- 2. Custom Validator untuk Username ---
		v.RegisterValidation("uniqueUsername", func(fl validator.FieldLevel) bool {
			username := fl.Field().String()
			var user models.User
			db := database.GetDB()
			err := db.Where("username = ?", username).First(&user).Error
			return gorm.ErrRecordNotFound == err
		})
	}
}
