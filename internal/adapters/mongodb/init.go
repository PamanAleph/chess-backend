// Package mongodb provides MongoDB adapter implementations.
// This is part of the Adapters layer in Hexagonal Architecture.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds MongoDB configuration
type Config struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// NewClient creates a new MongoDB client with the given configuration
func NewClient(config Config) (*mongo.Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetConnectTimeout(config.Timeout)
	clientOptions.SetServerSelectionTimeout(config.Timeout)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel = context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// Close closes the MongoDB client connection
func Close(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}

// CreateIndexes creates necessary indexes for the application
func CreateIndexes(client *mongo.Client, databaseName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := client.Database(databaseName)

	// Create indexes for users collection
	usersCollection := db.Collection("users")
	usernameIndex := mongo.IndexModel{
		Keys:    map[string]interface{}{"username": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := usersCollection.Indexes().CreateOne(ctx, usernameIndex)
	if err != nil {
		return fmt.Errorf("failed to create username index: %w", err)
	}

	// Create indexes for games collection
	gamesCollection := db.Collection("games")
	playerIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"white_player_id": 1},
		},
		{
			Keys: map[string]interface{}{"black_player_id": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
		{
			Keys: map[string]interface{}{"created_at": -1},
		},
	}
	_, err = gamesCollection.Indexes().CreateMany(ctx, playerIndexes)
	if err != nil {
		return fmt.Errorf("failed to create game indexes: %w", err)
	}

	return nil
}