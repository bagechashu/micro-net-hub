package internal

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AuthSession represents a user authentication session
type AuthSession struct {
	ID        string
	Username  string
	IsAdmin   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

// AuthSessionStore manages user authentication sessions
type AuthSessionStore struct {
	sessions map[string]*AuthSession
	mu       sync.RWMutex
	timeout  time.Duration
}

// NewAuthSessionStore creates a new session store
func NewAuthSessionStore(timeout time.Duration) *AuthSessionStore {
	store := &AuthSessionStore{
		sessions: make(map[string]*AuthSession),
		timeout:  timeout,
	}

	// Start cleanup goroutine to remove expired sessions
	go store.cleanupExpiredSessions()

	return store
}

// CreateSession creates a new session for a user
func (ss *AuthSessionStore) CreateSession(username string, isAdmin bool) (*AuthSession, error) {
	// Generate random session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &AuthSession{
		ID:        sessionID,
		Username:  username,
		IsAdmin:   isAdmin,
		CreatedAt: now,
		ExpiresAt: now.Add(ss.timeout),
	}

	ss.mu.Lock()
	ss.sessions[sessionID] = session
	ss.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
func (ss *AuthSessionStore) GetSession(sessionID string) (*AuthSession, error) {
	ss.mu.RLock()
	session, exists := ss.sessions[sessionID]
	ss.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		ss.mu.Lock()
		delete(ss.sessions, sessionID)
		ss.mu.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// ValidateSession validates a session and returns whether it's valid
func (ss *AuthSessionStore) ValidateSession(sessionID string) bool {
	_, err := ss.GetSession(sessionID)
	return err == nil
}

// RenewSession extends the expiration time of a session
func (ss *AuthSessionStore) RenewSession(sessionID string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		delete(ss.sessions, sessionID)
		return fmt.Errorf("session expired")
	}

	// Extend expiration
	session.ExpiresAt = time.Now().Add(ss.timeout)
	return nil
}

// DeleteSession removes a session
func (ss *AuthSessionStore) DeleteSession(sessionID string) (username string, err error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session not found")
	}
	username = session.Username
	delete(ss.sessions, sessionID)

	return username, nil
}

// cleanupExpiredSessions periodically removes expired sessions
func (ss *AuthSessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ss.mu.Lock()
		now := time.Now()
		count := 0

		for sessionID, session := range ss.sessions {
			if now.After(session.ExpiresAt) {
				delete(ss.sessions, sessionID)
				count++
			}
		}

		ss.mu.Unlock()

		if count > 0 {
			// Uncomment for debugging
			// log.Printf("[session] 清理了 %d 个过期会话", count)
		}
	}
}

// generateSessionID generates a random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetAllSessions returns all active sessions (for debugging/admin purposes)
func (ss *AuthSessionStore) GetAllSessions() []*AuthSession {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	sessions := make([]*AuthSession, 0, len(ss.sessions))
	now := time.Now()

	for _, session := range ss.sessions {
		if now.Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}
