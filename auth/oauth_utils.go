package auth

import (
	"context"
	"fmt"
	"time"
)

// CreateSessionForOAuthUser creates a session for a user who authenticated via OAuth
// without requiring password verification
func (p *DefaultAuthProvider) CreateSessionForOAuthUser(ctx context.Context, user *User) (*Session, error) {
	// Verify this is indeed an OAuth user
	if user.OAuthProvider == "" || user.OAuthID == "" {
		return nil, fmt.Errorf("user is not authenticated via OAuth")
	}

	// Update last login time
	user.LastLogin = p.getTime()
	if err := p.userStore.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user login time: %w", err)
	}

	// Create a new session directly
	session, err := p.createSession(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// For testing, allows overriding the current time
func (p *DefaultAuthProvider) getTime() time.Time {
	return time.Now()
}
