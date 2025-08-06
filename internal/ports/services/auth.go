// Package services defines the interfaces for business logic services.
// This is part of the Ports layer in Hexagonal Architecture.
// These interfaces define the contracts for application services.
package services

import (
	"context"

	"chess-backend/internal/domain/user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SignupRequest represents the data needed for user registration
type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SigninRequest represents the data needed for user login
type SigninRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response after authentication operations
type AuthResponse struct {
	Message   string `json:"message"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	User      *user.User `json:"user,omitempty"`
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// Signup registers a new user and creates a session
	Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error)

	// Signin authenticates a user and creates a session
	Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error)

	// Logout invalidates a user session
	Logout(ctx context.Context, sessionID string) error

	// GetCurrentUser retrieves user information from session
	GetCurrentUser(ctx context.Context, sessionID string) (*user.User, error)

	// RefreshSession extends the TTL of a session
	RefreshSession(ctx context.Context, sessionID string) error

	// ValidateSession checks if a session is valid and returns user ID
	ValidateSession(ctx context.Context, sessionID string) (primitive.ObjectID, error)

	// ChangePassword allows a user to change their password
	ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, userID primitive.ObjectID) (*user.User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *user.User) error

	// DeleteUser removes a user and all their sessions
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
}