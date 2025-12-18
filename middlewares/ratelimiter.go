package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"mygram-api/database"
	"mygram-api/dto"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig mengembalikan middleware untuk membatasi request
// limit: Batas request yang diizinkan (e.g., 10)
// window: Jendela waktu dalam detik (e.g., 60)
func RateLimiterConfig(limit int64, window time.Duration) gin.HandlerFunc {
	rdb := database.GetRedis()
	ctx := context.Background()

	if rdb == nil {
		log.Println("WARNING: Redis client is NIL. Rate Limiter is disabled.")
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		// Dapatkan Identifier (ID User jika terautentikasi, IP jika tidak)
		var identifier string

		// Coba ambil ID User dari JWT context (jika sudah melewati Auth)
		userData, exists := c.Get("userData")
		if exists {
			// Asumsi ID User ada di JWT claims
			if claims, ok := userData.(map[string]any); ok {
				if id, idOk := claims["id"].(string); idOk {
					identifier = "user:" + id
				}
			}
		}

		// Jika belum terautentikasi, gunakan IP Address sebagai identifier
		if identifier == "" {
			identifier = "ip:" + c.ClientIP()
		}

		// Key Redis: identifier:endpoint (e.g., user:uuid-123:comments)
		key := fmt.Sprintf("rate:%s:%s", identifier, c.Request.URL.Path)

		// 1. Increment counter dan simpan hasil increment
		// Menggunakan INCR: operasi atomic (aman untuk concurrent access)
		count, err := rdb.Incr(ctx, key).Result()

		if err != nil && err != redis.Nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.BaseResponseError{
				Success: false,
				Message: "Rate limiter internal error",
			})
			return
		}

		// 2. Jika ini request pertama (count == 1), set expiration time
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		// 3. Periksa Batas
		if count > limit {
			// Ambil waktu sisa (TTL) untuk dikembalikan ke client
			ttl := rdb.TTL(ctx, key).Val()

			// Set Retry-After header
			c.Writer.Header().Set("Retry-After", strconv.FormatInt(int64(ttl/time.Second), 10))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.BaseResponseError{
				Success: false,
				Message: fmt.Sprintf("Rate limit exceeded. You are limited to %d requests per %s. Try again in %d seconds.", limit, window, int(ttl.Seconds())),
			})
			return
		}

		c.Next()
	}
}
