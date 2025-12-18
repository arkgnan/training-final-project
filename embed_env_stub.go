//go:build !embed
// +build !embed

package main

import (
	"embed"
	"log"
)

// stub declaration so main.go can reference embeddedFiles even when not embedding
var embeddedFiles embed.FS

func loadEmbeddedConfig() {
	// No-op or minimal logging; environment variables expected from OS/env provider
	log.Println("Embedded .env not included in this build; using OS environment.")
}
