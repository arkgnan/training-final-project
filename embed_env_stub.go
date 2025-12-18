//go:build !embed
// +build !embed

package main

import "embed"

// stub declaration so main.go can reference embeddedFiles even when not embedding
var embeddedFiles embed.FS
