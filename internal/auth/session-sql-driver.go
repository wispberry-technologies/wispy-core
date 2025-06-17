package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type SqlSessionDriver struct {
	db     *sql.DB
	Config struct {
		SessionCookieName     string
		SessionCookieSameSite http.SameSite
		SessionTimeout        time.Duration
		SectionCookieMaxAge   int
		IsCookieSecure        bool
	}
}

// NewSqlSessionRepository creates a new session repository
func NewSessionSqlDriver(db *sql.DB) *SqlSessionDriver {
	return &SqlSessionDriver{db: db}
}

// CreateSession creates a new session for a user
func (s *SqlSessionDriver) CreateSession(userID, ipAddress, userAgent string) (*Session, error) {
	session := NewSession(userID, s.Config.SessionTimeout, ipAddress, userAgent)
	_, err := s.db.Exec(InsertSessionSQL,
		session.ID, session.UserID, session.Token, session.ExpiresAt,
		session.CreatedAt, session.UpdatedAt, session.IPAddress, session.UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

// GetSession retrieves a session by token
func (s *SqlSessionDriver) GetSession(token string) (*Session, error) {
	var session Session
	err := s.db.QueryRow(GetSessionByTokenSQL, token).Scan(
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
func (s *SqlSessionDriver) RefreshSession(sessionID string) error {
	newExpiry := time.Now().Add(s.Config.SessionTimeout)
	_, err := s.db.Exec(UpdateSessionSQL, newExpiry, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}
	return nil
}

// DeleteSession removes a session
func (s *SqlSessionDriver) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(DeleteSessionSQL, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteSessionByToken removes a session by token
func (s *SqlSessionDriver) DeleteSessionByToken(token string) error {
	_, err := s.db.Exec(DeleteSessionByTokenSQL, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user
func (s *SqlSessionDriver) DeleteUserSessions(userID string) error {
	_, err := s.db.Exec(DeleteUserSessionsSQL, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

// DeleteAllUserSessions deletes all sessions for a specific user
func (s *SqlSessionDriver) DeleteAllUserSessions(userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SqlSessionDriver) CleanupExpiredSessions() error {
	_, err := s.db.Exec(DeleteExpiredSessionsSQL)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// GetSessionFromRequest extracts session token from request cookie
func (s *SqlSessionDriver) GetSessionFromRequest(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie found")
	}
	session, err := s.GetSession(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}
	if session.IsExpired() {
		// Clean up expired session
		s.DeleteSession(session.ID)
		return nil, fmt.Errorf("session expired")
	}
	return session, nil
}

// SetSessionCookie sets a secure session cookie
func (s *SqlSessionDriver) SetSessionCookie(w http.ResponseWriter, token string) {
	var sameSite = http.SameSiteLaxMode
	if s.Config.SessionCookieSameSite != http.SameSiteDefaultMode {
		sameSite = s.Config.SessionCookieSameSite
	}

	cookieName := s.Config.SessionCookieName
	if cookieName == "" {
		cookieName = "wispy_auth_session"
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.Config.SectionCookieMaxAge),
		HttpOnly: true,
		Secure:   s.Config.IsCookieSecure,
		SameSite: sameSite,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (s *SqlSessionDriver) ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookieName := s.Config.SessionCookieName
	if cookieName == "" {
		cookieName = "wispy_auth_session"
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: s.Config.SessionCookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// SessionContext keys for storing session data in request context
type contextKey string

const (
	SessionContextKey contextKey = "session"
	UserContextKey    contextKey = "user"
)

// GetSessionFromContext retrieves session from context
func GetSessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(SessionContextKey).(*Session)
	return session, ok
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}

// SetSessionInContext stores session in context
func SetSessionInContext(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

// SetUserInContext stores user in context
func SetUserInContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
