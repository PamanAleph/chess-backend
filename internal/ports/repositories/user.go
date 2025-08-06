// Package repositories defines the interfaces for data persistence.
// This is part of the Ports layer in Hexagonal Architecture.
// Ports define contracts that adapters must implement.
package repositories

import (
	"context"

	"chess-backend/internal/domain/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	// Save creates a new user in the repository
	Save(ctx context.Context, user *user.User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*user.User, error)

	// Update updates an existing user in the repository
	Update(ctx context.Context, user *user.User) error

	// Delete removes a user from the repository
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Exists checks if a user exists by ID
	Exists(ctx context.Context, id primitive.ObjectID) (bool, error)

	// ExistsByUsername checks if a user exists by username
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// List retrieves users with pagination
	List(ctx context.Context, offset, limit int) ([]*user.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}