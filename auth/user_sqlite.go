package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteUserStore implements UserStore for SQLite
type SQLiteUserStore struct {
	db *sql.DB
}

// NewSQLiteUserStore creates a new SQLite user store
func NewSQLiteUserStore(db *sql.DB) (*SQLiteUserStore, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	store := &SQLiteUserStore{db: db}
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create user tables: %w", err)
	}

	return store, nil
}

// createTables ensures the necessary tables exist
func (s *SQLiteUserStore) createTables() error {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		display_name TEXT,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_login TIMESTAMP,
		oauth_provider TEXT,
		oauth_id TEXT,
		metadata BLOB
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`

	roleTable := `
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, role),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
	`

	_, err := s.db.Exec(userTable)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(roleTable)
	return err
}

// CreateUser implements UserStore.CreateUser
func (s *SQLiteUserStore) CreateUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		user.ID = GenerateID() // Implement GenerateID() elsewhere
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Hash the password if provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	// Serialize metadata to JSON
	var metadata []byte
	var err error
	if len(user.Metadata) > 0 {
		metadata = user.Metadata
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Will be a no-op if transaction succeeds

	// Insert user record
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (
			id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Email, user.Username, user.DisplayName, user.Password,
		user.CreatedAt, user.UpdatedAt, user.LastLogin,
		user.OAuthProvider, user.OAuthID, metadata)

	if err != nil {
		return err
	}

	// Insert user roles if any
	if len(user.Roles) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO user_roles (user_id, role) VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, role := range user.Roles {
			_, err = stmt.ExecContext(ctx, user.ID, role)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetUserByID implements UserStore.GetUserByID
func (s *SQLiteUserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		FROM users WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.OAuthProvider, &user.OAuthID, &metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Metadata = metadata

	// Get user roles
	roles, err := s.GetUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// GetUserByEmail implements UserStore.GetUserByEmail
func (s *SQLiteUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		FROM users WHERE email = ?
	`, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.OAuthProvider, &user.OAuthID, &metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Metadata = metadata

	// Get user roles
	roles, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// GetUserByUsername implements UserStore.GetUserByUsername
func (s *SQLiteUserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.OAuthProvider, &user.OAuthID, &metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Metadata = metadata

	// Get user roles
	roles, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// UpdateUser implements UserStore.UpdateUser
func (s *SQLiteUserStore) UpdateUser(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update user record
	_, err = tx.ExecContext(ctx, `
		UPDATE users SET 
			email = ?, 
			username = ?, 
			display_name = ?,
			updated_at = ?,
			last_login = ?,
			oauth_provider = ?,
			oauth_id = ?,
			metadata = ?
		WHERE id = ?
	`,
		user.Email,
		user.Username,
		user.DisplayName,
		user.UpdatedAt,
		user.LastLogin,
		user.OAuthProvider,
		user.OAuthID,
		user.Metadata,
		user.ID,
	)

	if err != nil {
		return err
	}

	// Clear existing roles and insert new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = ?", user.ID)
	if err != nil {
		return err
	}

	if len(user.Roles) > 0 {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO user_roles (user_id, role) VALUES (?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, role := range user.Roles {
			_, err = stmt.ExecContext(ctx, user.ID, role)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// DeleteUser implements UserStore.DeleteUser
func (s *SQLiteUserStore) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// ListUsers implements UserStore.ListUsers
func (s *SQLiteUserStore) ListUsers(ctx context.Context, offset, limit int) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user := &User{}
		var metadata []byte

		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.Password,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
			&user.OAuthProvider, &user.OAuthID, &metadata,
		)

		if err != nil {
			return nil, err
		}

		user.Metadata = metadata

		// Get user roles
		roles, err := s.GetUserRoles(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		user.Roles = roles

		users = append(users, user)
	}

	return users, nil
}

// CountUsers implements UserStore.CountUsers
func (s *SQLiteUserStore) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// VerifyPassword implements UserStore.VerifyPassword
func (s *SQLiteUserStore) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	// Get user's password and OAuth provider info
	var hashedPassword string
	var oauthProvider string
	err := s.db.QueryRowContext(ctx,
		"SELECT password, oauth_provider FROM users WHERE id = ?",
		userID).Scan(&hashedPassword, &oauthProvider)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user not found")
		}
		return false, err
	}

	// If the user was created via OAuth and has no password,
	// we can't verify with password - return an appropriate error
	if hashedPassword == "" && oauthProvider != "" {
		return false, errors.New("oauth user has no password")
	}

	// Regular password verification
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, nil
}

// UpdatePassword implements UserStore.UpdatePassword
func (s *SQLiteUserStore) UpdatePassword(ctx context.Context, userID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), userID)
	return err
}

// AddUserToRole implements UserStore.AddUserToRole
func (s *SQLiteUserStore) AddUserToRole(ctx context.Context, userID, role string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO user_roles (user_id, role) VALUES (?, ?)",
		userID, role)
	return err
}

// RemoveUserFromRole implements UserStore.RemoveUserFromRole
func (s *SQLiteUserStore) RemoveUserFromRole(ctx context.Context, userID, role string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM user_roles WHERE user_id = ? AND role = ?",
		userID, role)
	return err
}

// GetUserRoles implements UserStore.GetUserRoles
func (s *SQLiteUserStore) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT role FROM user_roles WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserByOAuthID implements UserStore.GetUserByOAuthID
func (s *SQLiteUserStore) GetUserByOAuthID(ctx context.Context, provider, oauthID string) (*User, error) {
	user := &User{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, display_name, password, 
			created_at, updated_at, last_login,
			oauth_provider, oauth_id, metadata
		FROM users WHERE oauth_provider = ? AND oauth_id = ?
	`, provider, oauthID).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.OAuthProvider, &user.OAuthID, &metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Metadata = metadata

	// Get user roles
	roles, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// LinkOAuthToUser implements UserStore.LinkOAuthToUser
func (s *SQLiteUserStore) LinkOAuthToUser(ctx context.Context, userID, provider, oauthID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET oauth_provider = ?, oauth_id = ? WHERE id = ?",
		provider, oauthID, userID)
	return err
}

// UnlinkOAuthFromUser implements UserStore.UnlinkOAuthFromUser
func (s *SQLiteUserStore) UnlinkOAuthFromUser(ctx context.Context, userID, provider string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET oauth_provider = NULL, oauth_id = NULL WHERE id = ? AND oauth_provider = ?",
		userID, provider)
	return err
}

// Helper function for generating a unique ID
func GenerateID() string {
	// In production you would use something like:
	// return uuid.New().String()
	// For this example, we'll use a timestamp-based ID
	return fmt.Sprintf("usr_%d", time.Now().UnixNano())
}
