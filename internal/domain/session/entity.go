// Package session contains the Session domain entity and its business logic.
// This is part of the Domain layer in Hexagonal Architecture.
// Domain entities should be pure business logic without any external dependencies.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session represents a user session in the domain
type Session struct {
	ID        string             `json:"id"`
	UserID    primitive.ObjectID `json:"user_id"`
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
	TTL       time.Duration      `json:"ttl"` // TTL as duration
}

// NewSession creates a new session for a user
func NewSession(userID primitive.ObjectID, ttl time.Duration) *Session {
	if userID.IsZero() || ttl <= 0 {
		panic("invalid session parameters")
	}

	sessionID, err := generateSessionID()
	if err != nil {
		panic("failed to generate session ID: " + err.Error())
	}

	now := time.Now()
	return &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		TTL:       ttl,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Refresh extends the session expiration time
func (s *Session) Refresh(ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	s.ExpiresAt = time.Now().Add(ttl)
	s.TTL = ttl
}

// IsValid checks if the session is valid
func (s *Session) IsValid() bool {
	return s.ID != "" && !s.UserID.IsZero() && s.TTL > 0
}

// GetRemainingTTL returns the remaining TTL as duration
func (s *Session) GetRemainingTTL() time.Duration {
	remaining := time.Until(s.ExpiresAt)
	if remaining <= 0 {
		return 0
	}
	return remaining
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}