//go:build embed
// +build embed

package main

import (
	"embed"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

//go:embed .env
var embeddedFiles embed.FS

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
