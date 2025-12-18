//go:build embed
// +build embed

package main

import "embed"

//go:embed .env
var embeddedFiles embed.FS
