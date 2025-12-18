package main

import (
	"fmt"
	"log"
	"mygram-api/database"
	"mygram-api/helpers"
	"mygram-api/router"
	"os"
	"strings"
)

var isReleaseBuild = "no"

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

	// Initialize email templates at startup so template errors are detected early.
	if err := helpers.InitEmailTemplates(); err != nil {
		log.Printf("Warning: failed to initialize email templates: %v. Templated emails will fallback to plain-text.", err)
	}

	// Setup and run the router
	r := router.SetupRouter()
	// Use PLATFORM PORT first (Railway sets PORT), then APP_PORT, then default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
		if port == "" {
			port = "8080"
		}
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
