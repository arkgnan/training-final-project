package database

import (
	"log"
	"mygram-api/models"

	"gorm.io/gorm"
)

// MigrateTables menjalankan semua proses AutoMigrate untuk model aplikasi
func MigrateTables(db *gorm.DB) {
	// Pastikan ekstensi uuid-ossp sudah diaktifkan jika menggunakan PostgreSQL
	// Ini penting untuk fungsi uuid_generate_v4() yang digunakan oleh Gorm.
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	if err != nil {
		log.Fatalf("Failed to create uuid-ossp extension: %v", err)
	}

	// Lakukan AutoMigrate untuk setiap model
	// Catatan: Gorm AutoMigrate tidak menghapus kolom atau tabel,
	// hanya menambahkan yang baru atau memodifikasi yang sesuai.
	db.AutoMigrate(
		&models.User{},
		&models.Photo{},
		&models.Comment{},
		&models.SocialMedia{},
	)

	log.Println("Database migration completed successfully!")
}
