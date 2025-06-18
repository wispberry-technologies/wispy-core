package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"wispy-core/pkg/common"
)

type SqlSessionDriver struct {
	db     *sql.DB
	Config SqlSessionDriverConfig
}

type SqlSessionDriverConfig struct {
	SessionCookieSameSite http.SameSite
	SessionCookieName     string
	SectionCookieMaxAge   time.Duration
	SessionTimeout        time.Duration
	IsCookieSecure        bool
}

// NewSqlSessionRepository creates a new SqlSessionDriver with the provided database connection
func NewSessionSqlDriver(db *sql.DB) *SqlSessionDriver {
	return &SqlSessionDriver{
		db: db,
		Config: SqlSessionDriverConfig{
			SessionCookieSameSite: http.SameSiteLaxMode,
			SessionCookieName:     "session",
			SectionCookieMaxAge:   7 * time.Hour * 24, // 7 days
			SessionTimeout:        7 * time.Hour * 24,
			IsCookieSecure:        false,
		},
	}
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
	common.Debug("Session cookie found: %s", cookie)
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
	cookie := &http.Cookie{
		Name:     s.Config.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.Config.SessionTimeout.Seconds()),
		HttpOnly: true,
		Secure:   s.Config.IsCookieSecure,
		SameSite: s.Config.SessionCookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (s *SqlSessionDriver) ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     s.Config.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: s.Config.SessionCookieSameSite,
	}
	http.SetCookie(w, cookie)
}
