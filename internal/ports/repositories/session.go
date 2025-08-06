// Package repositories defines the interfaces for data persistence.
// This is part of the Ports layer in Hexagonal Architecture.
// Ports define contracts that adapters must implement.
package repositories

import (
	"context"
	"time"

	"chess-backend/internal/domain/session"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SessionRepository defines the interface for session data persistence
type SessionRepository interface {
	// Save stores a session in the repository with TTL
	Save(ctx context.Context, session *session.Session) error

	// FindByID retrieves a session by its ID
	FindByID(ctx context.Context, sessionID string) (*session.Session, error)

	// FindByUserID retrieves all active sessions for a user
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*session.Session, error)

	// Delete removes a session from the repository
	Delete(ctx context.Context, sessionID string) error

	// DeleteByUserID removes all sessions for a specific user
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error

	// Refresh extends the TTL of a session
	Refresh(ctx context.Context, sessionID string, ttl time.Duration) error

	// Exists checks if a session exists and is not expired
	Exists(ctx context.Context, sessionID string) (bool, error)

	// GetUserID retrieves the user ID associated with a session
	GetUserID(ctx context.Context, sessionID string) (primitive.ObjectID, error)

	// Cleanup removes all expired sessions (for maintenance)
	Cleanup(ctx context.Context) error

	// Count returns the total number of active sessions
	Count(ctx context.Context) (int64, error)
}