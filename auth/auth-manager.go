package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/wispberry-technologies/wispy-core/models"
)

// AuthManager coordinates all authentication operations for a site
type AuthManager struct {
	db         *sql.DB
	repository *UserRepository
	session    *SessionManager
	password   *PasswordHasher
	config     *AuthConfig
	siteDomain string
}

// createTables creates all necessary authentication tables
func createTables(db *sql.DB) error {
	// Create users table
	if _, err := db.Exec(models.CreateUserTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sessions table
	if _, err := db.Exec(models.CreateSessionTableSQL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create OAuth accounts table
	if _, err := db.Exec(models.CreateOAuthAccountTableSQL); err != nil {
		return fmt.Errorf("failed to create oauth accounts table: %w", err)
	}

	return nil
}

// Register creates a new user account
func (am *AuthManager) Register(email, password, firstName, lastName, displayName string) (*models.User, error) {
	// Check if email already exists
	exists, err := am.repository.EmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate password
	if !IsValidPassword(password) {
		return nil, fmt.Errorf("password does not meet requirements")
	}

	// Hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.NewUser(email, firstName, lastName, displayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.PasswordHash = passwordHash

	if err := am.repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(email, password, ipAddress, userAgent string) (*models.User, *models.Session, error) {
	// Get user by email
	user, err := am.repository.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if am.repository.IsUserLocked(user) {
		return nil, nil, fmt.Errorf("account is locked")
	}

	// Verify password
	if err := VerifyPassword(password, user.PasswordHash); err != nil {
		// Increment failed login count
		user.FailedLoginCount++

		// Lock account if too many failed attempts
		if user.FailedLoginCount >= am.config.GetMaxFailedLoginAttempts() {
			user.IsLocked = true
			lockUntil := time.Now().Add(am.config.GetAccountLockoutDuration())
			user.LockedUntil = &lockUntil
		}

		am.repository.UpdateUserLoginAttempt(user.ID, user.FailedLoginCount, user.LockedUntil, nil)
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	now := time.Now()
	user.FailedLoginCount = 0
	user.LastLoginAt = &now
	user.IsLocked = false
	user.LockedUntil = nil

	if err := am.repository.UpdateUserLoginAttempt(user.ID, 0, nil, &now); err != nil {
		return nil, nil, fmt.Errorf("failed to update login info: %w", err)
	}

	// Create session
	session, err := am.session.CreateSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, session, nil
}

// Logout terminates a user's session
func (am *AuthManager) Logout(sessionToken string) error {
	return am.session.DeleteSessionByToken(sessionToken)
}

// LogoutAll terminates all sessions for a user
func (am *AuthManager) LogoutAll(userID string) error {
	return am.session.DeleteUserSessions(userID)
}

// ValidateSession checks if a session is valid and returns the user
func (am *AuthManager) ValidateSession(sessionToken string) (*models.User, *models.Session, error) {
	session, err := am.session.GetSession(sessionToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		am.session.DeleteSession(session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := am.repository.GetUserByID(session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("user account is inactive")
	}

	return user, session, nil
}

// RefreshSession extends the session expiration time
func (am *AuthManager) RefreshSession(sessionID string) error {
	return am.session.RefreshSession(sessionID)
}

// ChangePassword changes a user's password
func (am *AuthManager) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := am.repository.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := VerifyPassword(currentPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if !IsValidPassword(newPassword) {
		return fmt.Errorf("new password does not meet requirements")
	}

	// Hash new password
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.PasswordHash = newPasswordHash
	if err := am.repository.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all existing sessions except current one
	return am.session.DeleteUserSessions(userID)
}

// GetUser retrieves a user by ID
func (am *AuthManager) GetUser(userID string) (*models.User, error) {
	return am.repository.GetUserByID(userID)
}

// UpdateUser updates user information
func (am *AuthManager) UpdateUser(user *models.User) error {
	return am.repository.UpdateUser(user)
}

// CleanupExpiredSessions removes expired sessions
func (am *AuthManager) CleanupExpiredSessions() error {
	return am.session.CleanupExpiredSessions()
}

// SetSessionCookie sets a session cookie in the response
func (am *AuthManager) SetSessionCookie(w http.ResponseWriter, token string) {
	am.session.SetSessionCookie(w, token)
}

// ClearSessionCookie clears the session cookie
func (am *AuthManager) ClearSessionCookie(w http.ResponseWriter) {
	am.session.ClearSessionCookie(w)
}

// GetSessionFromRequest extracts session from HTTP request
func (am *AuthManager) GetSessionFromRequest(r *http.Request) (*models.Session, error) {
	return am.session.GetSessionFromRequest(r)
}

// GetUserRepository returns the user repository for direct access
func (am *AuthManager) GetUserRepository() *UserRepository {
	return am.repository
}

// Close closes the database connection
func (am *AuthManager) Close() error {
	return am.db.Close()
}
