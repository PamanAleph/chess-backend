package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Hexagonal Architecture imports
	httpAdapter "chess-backend/internal/adapters/http"
	"chess-backend/internal/adapters/mongodb"
	"chess-backend/internal/adapters/redis"
	"chess-backend/internal/application/auth"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("Warning: Could not load .env.local file: %v", err)
		// Tidak fatal, bisa menggunakan environment variables sistem
	}

	// Initialize MongoDB client
	mongoConfig := mongodb.Config{
		URI:      getEnv("MONGODB_URI", ""),
		Database: getEnv("MONGODB_DATABASE", ""),
		Timeout:  10 * time.Second,
	}
	mongoClient, err := mongodb.NewClient(mongoConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if closeErr := mongodb.Close(mongoClient); closeErr != nil {
			log.Printf("Error closing MongoDB connection: %v", closeErr)
		}
	}()

	// Initialize Redis client
	redisConfig := redis.Config{
		Addr:     getEnv("REDIS_ADDR", ""),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
		Timeout:  5 * time.Second,
	}
	redisClient, err := redis.NewClient(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		if err := redis.Close(redisClient); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// Initialize repositories (Adapters)
	userRepo := mongodb.NewUserRepository(mongoClient, mongoConfig.Database)
	sessionRepo := redis.NewSessionRepository(redisClient)

	// Initialize application services
	authService := auth.NewAuthService(userRepo, sessionRepo)

	// Initialize HTTP server with dependency injection
	server := httpAdapter.NewServer(authService, nil) // gameService will be added later
	router := server.GetRouter()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Chess WebSocket Backend server starting on port %s", port)
		log.Printf("Environment: %s", os.Getenv("ENVIRONMENT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
