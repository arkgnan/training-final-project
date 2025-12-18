package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig mengembalikan konfigurasi middleware CORS
func CORSConfig() gin.HandlerFunc {
	return cors.New(cors.Config{
		// Jika bukan open api, ganti "*" dengan daftar domain frontend spesifik Anda (e.g., "https://mygram.com")
		AllowOrigins: []string{"*"},

		// Izinkan method HTTP yang digunakan di API Anda
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},

		// Izinkan header yang mungkin digunakan (terutama Authorization untuk JWT)
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},

		// Apakah kredensial (seperti cookies) akan diizinkan (tidak relevan untuk JWT token di Header)
		AllowCredentials: true,

		// Maksimum waktu preflight request (OPTIONS) di-cache oleh browser
		MaxAge: 12 * time.Hour,
	})
}

func SSEProtocolHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Headers Wajib untuk Protokol SSE
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")

		// Header Penting untuk Proxy (misalnya Nginx/Cloudflare)
		ctx.Writer.Header().Set("X-Accel-Buffering", "no")

		ctx.Next()
	}
}
