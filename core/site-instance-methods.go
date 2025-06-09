package core

import (
	"fmt"
	"net/http"
	"time"

	"wispy-core/auth"
	"wispy-core/cache"
	"wispy-core/models"
)

type SiteInstance = models.SiteInstance

// Register creates a new user account for a site
func Register(instance *SiteInstance, email, password, firstName, lastName, displayName string) (*models.User, error) {
	db, err := cache.GetDB(instance, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection \"%s\": %w", "users", err)
	}

	repository := auth.NewUserSqlDriver(db)

	// Check if email already exists
	exists, err := repository.EmailExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate password
	if !auth.IsValidPassword(password) {
		return nil, fmt.Errorf("password does not meet requirements")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.NewUser(email, firstName, lastName, displayName)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.PasswordHash = passwordHash

	if err := repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and creates a session for a site
func Login(instance *SiteInstance, email, password, ipAddress, userAgent string) (*models.User, *models.Session, error) {
	db, err := cache.GetDB(instance, "users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	repository := auth.NewUserSqlDriver(db)

	// Get user by email
	user, err := repository.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if repository.IsUserLocked(user) {
		return nil, nil, fmt.Errorf("account is locked")
	}

	// Verify password
	if err := auth.VerifyPassword(password, user.PasswordHash); err != nil {
		// Increment failed login count
		user.FailedLoginCount++

		// Lock account if too many failed attempts
		if user.FailedLoginCount >= instance.Config.MaxFailedLoginAttempts {
			user.IsLocked = true
			lockUntil := time.Now().Add(instance.Config.FailedLoginAttemptLockDuration)
			user.LockedUntil = &lockUntil
		}

		repository.UpdateUserLoginAttempt(user.ID, user.FailedLoginCount, user.LockedUntil, nil)
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login count on successful login
	now := time.Now()
	user.FailedLoginCount = 0
	user.LastLoginAt = &now
	user.IsLocked = false
	user.LockedUntil = nil

	if err := repository.UpdateUserLoginAttempt(user.ID, 0, nil, &now); err != nil {
		return nil, nil, fmt.Errorf("failed to update login info: %w", err)
	}

	// Create session
	sessionDriver := auth.NewSessionSqlDriver(db)
	session, err := sessionDriver.CreateSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, session, nil
}

// ValidateSession checks if a session is valid and returns the user for a site
func ValidateSession(instance *SiteInstance, sessionToken string) (*models.User, *models.Session, error) {
	db, err := cache.GetDB(instance, "users")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sessionDriver := auth.NewSessionSqlDriver(db)
	repository := auth.NewUserSqlDriver(db)

	session, err := sessionDriver.GetSession(sessionToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session: %w", err)
	}

	if session.IsExpired() {
		sessionDriver.DeleteSession(session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := repository.GetUserByID(session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if user.IsLocked {
		return nil, nil, fmt.Errorf("account is locked")
	}

	return user, session, nil
}

// GetSessionFromRequest extracts session from request for a site
func GetSessionFromRequest(instance *SiteInstance, r *http.Request) (*models.Session, error) {
	db, err := cache.GetDB(instance, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sessionDriver := auth.NewSessionSqlDriver(db)

	return sessionDriver.GetSessionFromRequest(r)
}

// RequireAuth middleware that requires authentication for a site
func RequireAuth(instance *SiteInstance) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := GetSessionFromRequest(instance, r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			db, err := cache.GetDB(instance, "users")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			user, _, err := auth.ValidateSession(db, session.Token)
			if err != nil { // Clear invalid session cookie
				db, _ := cache.GetDB(instance, "users")
				sessionDriver := auth.NewSessionSqlDriver(db)
				sessionDriver.ClearSessionCookie(w, r)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user and session to request context
			ctx := auth.SetUserInContext(r.Context(), user)
			ctx = auth.SetSessionInContext(ctx, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles middleware that requires specific roles for a site
func RequireRoles(instance *SiteInstance, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return RequireAuth(instance)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			userRoles := user.GetRoles()
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}
