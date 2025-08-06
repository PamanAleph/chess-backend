// Package redis implements the session repository interface using Redis.
// This is part of the Adapters layer in Hexagonal Architecture.
// Adapters implement the ports defined in the ports layer.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"chess-backend/internal/domain/session"
	"chess-backend/internal/ports/repositories"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// sessionRepository implements the SessionRepository interface using Redis
type sessionRepository struct {
	client *redis.Client
	prefix string
}

// NewSessionRepository creates a new Redis session repository
func NewSessionRepository(client *redis.Client) repositories.SessionRepository {
	return &sessionRepository{
		client: client,
		prefix: "session:",
	}
}

// getKey returns the Redis key for a session
func (r *sessionRepository) getKey(sessionID string) string {
	return r.prefix + sessionID
}

// getUserSessionsKey returns the Redis key for user sessions set
func (r *sessionRepository) getUserSessionsKey(userID primitive.ObjectID) string {
	return fmt.Sprintf("user_sessions:%s", userID.Hex())
}

// Save stores a session in Redis with TTL
func (r *sessionRepository) Save(ctx context.Context, sess *session.Session) error {
	if sess == nil {
		return errors.New("session cannot be nil")
	}

	if sess.ID == "" || sess.UserID.IsZero() {
		return errors.New("invalid session data")
	}

	// Serialize session to JSON
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := r.getKey(sess.ID)
	userSessionsKey := r.getUserSessionsKey(sess.UserID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Calculate TTL duration
	ttl := time.Until(sess.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default TTL if expired
	}

	// Set session data with TTL
	pipe.Set(ctx, key, data, ttl)

	// Add session ID to user's sessions set
	pipe.SAdd(ctx, userSessionsKey, sess.ID)
	pipe.Expire(ctx, userSessionsKey, ttl)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// FindByID retrieves a session by its ID
func (r *sessionRepository) FindByID(ctx context.Context, sessionID string) (*session.Session, error) {
	key := r.getKey(sessionID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var sess session.Session
	err = json.Unmarshal([]byte(data), &sess)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if session is expired
	if sess.IsExpired() {
		// Clean up expired session
		r.Delete(ctx, sessionID)
		return nil, errors.New("session expired")
	}

	return &sess, nil
}

// FindByUserID retrieves all active sessions for a user
func (r *sessionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*session.Session, error) {
	userSessionsKey := r.getUserSessionsKey(userID)

	// Get all session IDs for the user
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*session.Session{}, nil
		}
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []*session.Session
	for _, sessionID := range sessionIDs {
		sess, err := r.FindByID(ctx, sessionID)
		if err != nil {
			// Remove invalid session from set
			r.client.SRem(ctx, userSessionsKey, sessionID)
			continue
		}
		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// Delete removes a session from Redis
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	// First get the session to find the user ID
	sess, err := r.FindByID(ctx, sessionID)
	if err != nil {
		// If session doesn't exist, consider it deleted
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}

	key := r.getKey(sessionID)
	userSessionsKey := r.getUserSessionsKey(sess.UserID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Delete session data
	pipe.Del(ctx, key)

	// Remove session ID from user's sessions set
	pipe.SRem(ctx, userSessionsKey, sessionID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteByUserID removes all sessions for a specific user
func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	userSessionsKey := r.getUserSessionsKey(userID)

	// Get all session IDs for the user
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Delete all session data
	for _, sessionID := range sessionIDs {
		key := r.getKey(sessionID)
		pipe.Del(ctx, key)
	}

	// Delete user sessions set
	pipe.Del(ctx, userSessionsKey)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// Refresh extends the TTL of a session
func (r *sessionRepository) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	// Get current session
	sess, err := r.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	key := r.getKey(sessionID)
	userSessionsKey := r.getUserSessionsKey(sess.UserID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Extend TTL for session data
	pipe.Expire(ctx, key, ttl)

	// Extend TTL for user sessions set
	pipe.Expire(ctx, userSessionsKey, ttl)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	return nil
}

// Exists checks if a session exists and is not expired
func (r *sessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := r.getKey(sessionID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists > 0, nil
}

// GetUserID retrieves the user ID associated with a session
func (r *sessionRepository) GetUserID(ctx context.Context, sessionID string) (primitive.ObjectID, error) {
	sess, err := r.FindByID(ctx, sessionID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return sess.UserID, nil
}

// Cleanup removes all expired sessions (for maintenance)
func (r *sessionRepository) Cleanup(ctx context.Context) error {
	// Redis automatically handles TTL expiration, so this is mainly for cleanup
	// of orphaned user session sets
	pattern := "user_sessions:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get user session keys: %w", err)
	}

	for _, key := range keys {
		// Get all session IDs in the set
		sessionIDs, err := r.client.SMembers(ctx, key).Result()
		if err != nil {
			continue
		}

		// Check each session and remove expired ones
		for _, sessionID := range sessionIDs {
			exists, existsErr := r.Exists(ctx, sessionID)
			if existsErr != nil || !exists {
				r.client.SRem(ctx, key, sessionID)
			}
		}

		// If set is empty, delete it
		count, err := r.client.SCard(ctx, key).Result()
		if err == nil && count == 0 {
			r.client.Del(ctx, key)
		}
	}

	return nil
}

// Count returns the total number of active sessions
func (r *sessionRepository) Count(ctx context.Context) (int64, error) {
	pattern := r.prefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return int64(len(keys)), nil
}