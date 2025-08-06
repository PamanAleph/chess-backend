// Package auth implements the authentication business logic.
// This is part of the Application layer in Hexagonal Architecture.
// Application layer orchestrates domain entities and coordinates with adapters.
package auth

import (
	"context"
	"errors"
	"time"

	"chess-backend/internal/domain/session"
	"chess-backend/internal/domain/user"
	"chess-backend/internal/ports/repositories"
	"chess-backend/internal/ports/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// authService implements the AuthService interface
type authService struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	sessionTTL  time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repositories.UserRepository, sessionRepo repositories.SessionRepository) services.AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionTTL:  24 * time.Hour, // Default 24 hours
	}
}

// Signup registers a new user and creates a session
func (s *authService) Signup(ctx context.Context, req services.SignupRequest) (*services.AuthResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	if len(req.Username) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}

	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// Check if username already exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Create new user
	newUser, err := user.NewUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// Save user to repository
	err = s.userRepo.Save(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// Create session for the new user
	newSession := session.NewSession(newUser.ID, s.sessionTTL)
	err = s.sessionRepo.Save(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return &services.AuthResponse{
		Message:   "User registered successfully",
		UserID:    newUser.ID.Hex(),
		SessionID: newSession.ID,
		User:      newUser,
	}, nil
}

// Signin authenticates a user and creates a session
func (s *authService) Signin(ctx context.Context, req services.SigninRequest) (*services.AuthResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	// Find user by username
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Validate password
	if !existingUser.ValidatePassword(req.Password) {
		return nil, errors.New("invalid username or password")
	}

	// Create new session
	newSession := session.NewSession(existingUser.ID, s.sessionTTL)
	err = s.sessionRepo.Save(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return &services.AuthResponse{
		Message:   "Login successful",
		UserID:    existingUser.ID.Hex(),
		SessionID: newSession.ID,
		User:      existingUser,
	}, nil
}

// Logout invalidates a user session
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	return s.sessionRepo.Delete(ctx, sessionID)
}

// GetCurrentUser retrieves user information from session
func (s *authService) GetCurrentUser(ctx context.Context, sessionID string) (*user.User, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	// Get session
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get user
	return s.userRepo.FindByID(ctx, sess.UserID)
}

// RefreshSession extends the TTL of a session
func (s *authService) RefreshSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	return s.sessionRepo.Refresh(ctx, sessionID, s.sessionTTL)
}

// ValidateSession checks if a session is valid and returns user ID
func (s *authService) ValidateSession(ctx context.Context, sessionID string) (primitive.ObjectID, error) {
	if sessionID == "" {
		return primitive.NilObjectID, errors.New("session ID is required")
	}

	// Get session
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Check if session is expired
	if sess.IsExpired() {
		// Clean up expired session
		s.sessionRepo.Delete(ctx, sessionID)
		return primitive.NilObjectID, errors.New("session expired")
	}

	return sess.UserID, nil
}

// ChangePassword allows a user to change their password
func (s *authService) ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return errors.New("old password and new password are required")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}

	// Get user
	existingUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Validate old password
	if !existingUser.ValidatePassword(oldPassword) {
		return errors.New("invalid old password")
	}

	// Update password
	err = existingUser.UpdatePassword(newPassword)
	if err != nil {
		return err
	}

	// Save updated user
	return s.userRepo.Update(ctx, existingUser)
}

// GetUserByID retrieves a user by their ID
func (s *authService) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*user.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// UpdateUser updates user information
func (s *authService) UpdateUser(ctx context.Context, user *user.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if !user.IsValid() {
		return errors.New("invalid user data")
	}

	return s.userRepo.Update(ctx, user)
}

// DeleteUser removes a user and all their sessions
func (s *authService) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	// Delete all user sessions first
	err := s.sessionRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Delete user
	return s.userRepo.Delete(ctx, userID)
}