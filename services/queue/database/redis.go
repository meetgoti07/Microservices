package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gin-quickstart/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis initializes the Redis connection
func InitRedis(cfg *config.Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// GetRedis returns the Redis client
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
