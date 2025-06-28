package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SQLiteSessionStore implements SessionStore for SQLite
type SQLiteSessionStore struct {
	db *sql.DB
}

// NewSQLiteSessionStore creates a new SQLite session store
func NewSQLiteSessionStore(db *sql.DB) (*SQLiteSessionStore, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	store := &SQLiteSessionStore{db: db}
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create session tables: %w", err)
	}

	return store, nil
}

// createTables ensures the necessary tables exist
func (s *SQLiteSessionStore) createTables() error {
	sessionTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		ip TEXT,
		user_agent TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		data BLOB,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	sessionDataTable := `
	CREATE TABLE IF NOT EXISTS session_data (
		session_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT,
		PRIMARY KEY (session_id, key),
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	`

	_, err := s.db.Exec(sessionTable)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(sessionDataTable)
	return err
}

// CreateSession implements SessionStore.CreateSession
func (s *SQLiteSessionStore) CreateSession(ctx context.Context, session *Session) error {
	if session.ID == "" {
		session.ID = fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}

	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (
			id, user_id, token, expires_at, 
			ip, user_agent, created_at, updated_at, data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.ID, session.UserID, session.Token, session.ExpiresAt,
		session.IP, session.UserAgent, session.CreatedAt, session.UpdatedAt, session.Data,
	)

	return err
}

// GetSessionByID implements SessionStore.GetSessionByID
func (s *SQLiteSessionStore) GetSessionByID(ctx context.Context, id string) (*Session, error) {
	session := &Session{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, 
			ip, user_agent, created_at, updated_at, data
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.IP, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt, &session.Data,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return session, nil
}

// GetSessionByToken implements SessionStore.GetSessionByToken
func (s *SQLiteSessionStore) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	session := &Session{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, 
			ip, user_agent, created_at, updated_at, data
		FROM sessions WHERE token = ?
	`, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.IP, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt, &session.Data,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return session, nil
}

// UpdateSession implements SessionStore.UpdateSession
func (s *SQLiteSessionStore) UpdateSession(ctx context.Context, session *Session) error {
	session.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx, `
		UPDATE sessions SET 
			user_id = ?,
			token = ?,
			expires_at = ?,
			ip = ?,
			user_agent = ?,
			updated_at = ?,
			data = ?
		WHERE id = ?
	`,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		session.IP,
		session.UserAgent,
		session.UpdatedAt,
		session.Data,
		session.ID,
	)

	return err
}

// DeleteSession implements SessionStore.DeleteSession
func (s *SQLiteSessionStore) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	return err
}

// GetUserSessions implements SessionStore.GetUserSessions
func (s *SQLiteSessionStore) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, token, expires_at, 
			ip, user_agent, created_at, updated_at, data
		FROM sessions WHERE user_id = ? ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session

	for rows.Next() {
		session := &Session{}

		err := rows.Scan(
			&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
			&session.IP, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt, &session.Data,
		)

		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteUserSessions implements SessionStore.DeleteUserSessions
func (s *SQLiteSessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// DeleteExpiredSessions implements SessionStore.DeleteExpiredSessions
func (s *SQLiteSessionStore) DeleteExpiredSessions(ctx context.Context) (int, error) {
	result, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	return int(affected), err
}

// StoreSessionData implements SessionStore.StoreSessionData
func (s *SQLiteSessionStore) StoreSessionData(ctx context.Context, sessionID string, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize session data: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO session_data (session_id, key, value) 
		VALUES (?, ?, ?) 
		ON CONFLICT (session_id, key) DO UPDATE SET value = ?
	`, sessionID, key, string(valueBytes), string(valueBytes))

	return err
}

// GetSessionData implements SessionStore.GetSessionData
func (s *SQLiteSessionStore) GetSessionData(ctx context.Context, sessionID string, key string, valuePtr interface{}) error {
	var valueStr string

	err := s.db.QueryRowContext(ctx, "SELECT value FROM session_data WHERE session_id = ? AND key = ?",
		sessionID, key).Scan(&valueStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("session data not found")
		}
		return err
	}

	return json.Unmarshal([]byte(valueStr), valuePtr)
}

// RemoveSessionData implements SessionStore.RemoveSessionData
func (s *SQLiteSessionStore) RemoveSessionData(ctx context.Context, sessionID string, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM session_data WHERE session_id = ? AND key = ?",
		sessionID, key)
	return err
}
