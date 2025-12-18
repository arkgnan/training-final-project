package main

import (
	"embed"
	"fmt"
	"log"
	"mygram-api/database"
	"mygram-api/helpers"
	"mygram-api/router"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

//go:embed .env
var embeddedFiles embed.FS
var isReleaseBuild = "no"

func loadEmbeddedConfig() {
	// Membaca konten file .env dari binary yang di-embed
	data, err := embeddedFiles.ReadFile(".env")
	if err != nil {
		// PENTING: Jika file .env tidak ditemukan, fallback ke environment OS (untuk production)
		log.Println("Warning: Could not read embedded .env file. Falling back to OS environment.")
		return
	}

	// Mengurai konten file .env dari string/byte dan memuatnya ke dalam environment OS
	reader := strings.NewReader(string(data))
	envMap, err := godotenv.Parse(reader)
	if err != nil {
		log.Fatalf("Error parsing embedded .env file: %v", err)
	}
	for key, value := range envMap {
		os.Setenv(key, value)
	}
	log.Println("Configuration loaded successfully from embedded .env file.")
}

// @title MyGram REST API
// @version 1.0
// @description This is the final project API for MyGram application.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if isReleaseBuild == "yes" {
		loadEmbeddedConfig()
	}
	// Inisialisasi Redis
	database.StartRedis()
	// Initialize DB connection and run migrations
	database.StartDB()
	helpers.RegisterCustomValidator()
	// Setup and run the router
	r := router.SetupRouter()
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // default port if not set
	}
	host := os.Getenv("APP_URL")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if host == "" {
		host = "127.0.0.1" // fallback to localhost
	}
	addr := host + ":" + port
	fmt.Printf("Listening %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
