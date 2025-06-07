package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/wispberry-technologies/wispy-core/models"
)

// SessionManager handles session management using secure cookies
type SessionManager struct {
	db     *sql.DB
	config *AuthConfig
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sql.DB, config *AuthConfig) *SessionManager {
	return &SessionManager{
		db:     db,
		config: config,
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, ipAddress, userAgent string) (*models.Session, error) {
	session := models.NewSession(userID, sm.config.GetSessionTimeout(), ipAddress, userAgent)

	_, err := sm.db.Exec(models.InsertSessionSQL,
		session.ID, session.UserID, session.Token, session.ExpiresAt,
		session.CreatedAt, session.UpdatedAt, session.IPAddress, session.UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by token
func (sm *SessionManager) GetSession(token string) (*models.Session, error) {
	var session models.Session

	err := sm.db.QueryRow(models.GetSessionByTokenSQL, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.CreatedAt, &session.UpdatedAt, &session.IPAddress, &session.UserAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// RefreshSession extends the session expiration time
func (sm *SessionManager) RefreshSession(sessionID string) error {
	newExpiry := time.Now().Add(sm.config.GetSessionTimeout())
	_, err := sm.db.Exec(models.UpdateSessionSQL, newExpiry, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}
	return nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) error {
	_, err := sm.db.Exec(models.DeleteSessionSQL, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteSessionByToken removes a session by token
func (sm *SessionManager) DeleteSessionByToken(token string) error {
	_, err := sm.db.Exec(models.DeleteSessionByTokenSQL, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user
func (sm *SessionManager) DeleteUserSessions(userID string) error {
	_, err := sm.db.Exec(models.DeleteUserSessionsSQL, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

// DeleteAllUserSessions deletes all sessions for a specific user
func (sm *SessionManager) DeleteAllUserSessions(userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := sm.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() error {
	_, err := sm.db.Exec(models.DeleteExpiredSessionsSQL)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// SetSessionCookie sets a secure session cookie
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     sm.config.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sm.config.GetSessionTimeout().Seconds()),
		HttpOnly: sm.config.SessionCookieHTTPOnly,
		Secure:   sm.config.SessionCookieSecure,
		SameSite: sm.getSameSiteAttribute(),
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.config.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: sm.config.SessionCookieHTTPOnly,
		Secure:   sm.config.SessionCookieSecure,
		SameSite: sm.getSameSiteAttribute(),
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest extracts session token from request cookie
func (sm *SessionManager) GetSessionFromRequest(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie(sm.config.SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie found")
	}

	session, err := sm.GetSession(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		// Clean up expired session
		sm.DeleteSession(session.ID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// getSameSiteAttribute converts string to http.SameSite
func (sm *SessionManager) getSameSiteAttribute() http.SameSite {
	switch sm.config.SessionCookieSameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}

// generateSessionToken generates a cryptographically secure session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SessionContext keys for storing session data in request context
type contextKey string

const (
	SessionContextKey contextKey = "session"
	UserContextKey    contextKey = "user"
)

// GetSessionFromContext retrieves session from context
func GetSessionFromContext(ctx context.Context) (*models.Session, bool) {
	session, ok := ctx.Value(SessionContextKey).(*models.Session)
	return session, ok
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// SetSessionInContext stores session in context
func SetSessionInContext(ctx context.Context, session *models.Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

// SetUserInContext stores user in context
func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
