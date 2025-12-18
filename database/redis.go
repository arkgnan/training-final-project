package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func StartRedis() {
	// Ambil konfigurasi dari environment (pastikan sudah ada di .env Anda)
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	if redisHost == "" {
		redisHost = "localhost" // Default jika tidak diset
	}
	if redisPort == "" {
		redisPort = "6379" // Default jika tidak diset
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPass,
		DB:       0, // Default DB
	})

	// Cek koneksi
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis!")
}

func GetRedis() *redis.Client {
	return rdb
}
