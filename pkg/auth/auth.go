package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"wispy-core/pkg/common"
)

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// Register creates a new user account for a site
func Register(db *sql.DB, domain, email, password, firstName, lastName, displayName string) (*User, error) {
	usersDriver := NewUserSqlDriver(db)

	// Check if email already exists
	exists, err := usersDriver.EmailExists(email)
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
	user := NewUser(email, firstName, lastName, displayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.PasswordHash = passwordHash

	if err := usersDriver.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and creates a session for a site
func Login(db *sql.DB, domain, email, password, ipAddress, userAgent string, maxAttempts int, lockDuration time.Duration) (*User, *Session, error) {
	usersDriver := NewUserSqlDriver(db)

	// Get user by email
	user, err := usersDriver.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if usersDriver.IsUserLocked(user) {
		return nil, nil, fmt.Errorf("account is locked")
	}

	// Verify password
	if err := VerifyPassword(password, user.PasswordHash); err != nil {
		// Increment failed login count
		user.FailedLoginCount++

		// Lock account if too many failed attempts
		if user.FailedLoginCount >= maxAttempts {
			user.IsLocked = true
			lockUntil := time.Now().Add(lockDuration)
			user.LockedUntil = &lockUntil
		}

		usersDriver.UpdateUserLoginAttempt(user.ID, user.FailedLoginCount, user.LockedUntil, nil)
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	now := time.Now()
	user.FailedLoginCount = 0
	user.LastLoginAt = &now
	user.IsLocked = false
	user.LockedUntil = nil

	if err := usersDriver.UpdateUserLoginAttempt(user.ID, 0, nil, &now); err != nil {
		return nil, nil, fmt.Errorf("failed to update login info: %w", err)
	}

	// Create session
	sessionDriver := NewSessionSqlDriver(db)
	session, err := sessionDriver.CreateSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, session, nil
}

// ValidateSession checks if a session is valid and returns the user for a site
func ValidateSession(db *sql.DB, sessionToken string) (*User, *Session, error) {
	sessionDriver := NewSessionSqlDriver(db)
	userDriver := NewUserSqlDriver(db)

	common.Debug("Validating session: %s", sessionToken)
	session, err := sessionDriver.GetSession(sessionToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		sessionDriver.DeleteSession(session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := userDriver.GetUserByID(session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if user.IsLocked {
		return nil, nil, fmt.Errorf("account is locked")
	}

	return user, session, nil
}

// GetSessionFromRequest extracts session from request for a site
func GetSessionFromRequest(db *sql.DB, r *http.Request) (*Session, error) {
	sessionDriver := NewSessionSqlDriver(db)

	return sessionDriver.GetSessionFromRequest(r)
}
