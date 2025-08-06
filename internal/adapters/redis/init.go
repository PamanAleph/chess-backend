// Package redis provides Redis adapter implementations.
// This is part of the Adapters layer in Hexagonal Architecture.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
	Timeout  time.Duration
}

// NewClient creates a new Redis client with the given configuration
func NewClient(config Config) (*redis.Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.Timeout,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// Close closes the Redis client connection
func Close(client *redis.Client) error {
	return client.Close()
}

// FlushDB clears all data from the current Redis database (use with caution)
func FlushDB(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.FlushDB(ctx).Err()
}

// GetInfo returns Redis server information
func GetInfo(client *redis.Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Info(ctx).Result()
}