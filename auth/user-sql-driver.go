package auth

import (
	"database/sql"
	"fmt"
	"time"

	"wispy-core/models"
)

// UserRepository handles database operations for users
type UserSqlDriver struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserSqlDriver(db *sql.DB) *UserSqlDriver {
	return &UserSqlDriver{db: db}
}

// CreateUser creates a new user in the database
func (this *UserSqlDriver) CreateUser(user *models.User) error {
	var database = this.db

	_, err := database.Exec(models.InsertUserSQL,
		user.ID, user.Email, user.EmailVerified, user.PasswordHash,
		user.FirstName, user.LastName, user.DisplayName, user.Avatar,
		user.Roles, user.IsActive, user.IsLocked, user.FailedLoginCount,
		user.TwoFactorEnabled, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by email address
func (this *UserSqlDriver) GetUserByEmail(email string) (*models.User, error) {
	var database = this.db

	var user models.User
	var emailVerifiedAt sql.NullTime
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime
	var twoFactorSecret sql.NullString

	err := database.QueryRow(models.GetUserByEmailSQL, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &emailVerifiedAt,
		&user.PasswordHash, &user.FirstName, &user.LastName, &user.DisplayName,
		&user.Avatar, &user.Roles, &user.IsActive, &user.IsLocked,
		&lockedUntil, &user.FailedLoginCount, &lastLoginAt,
		&user.TwoFactorEnabled, &twoFactorSecret, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert nullable fields
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if twoFactorSecret.Valid {
		user.TwoFactorSecret = twoFactorSecret.String
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (this *UserSqlDriver) GetUserByID(id string) (*models.User, error) {
	var database = this.db
	var user models.User
	var emailVerifiedAt sql.NullTime
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime
	var twoFactorSecret sql.NullString

	err := database.QueryRow(models.GetUserByIDSQL, id).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &emailVerifiedAt,
		&user.PasswordHash, &user.FirstName, &user.LastName, &user.DisplayName,
		&user.Avatar, &user.Roles, &user.IsActive, &user.IsLocked,
		&lockedUntil, &user.FailedLoginCount, &lastLoginAt,
		&user.TwoFactorEnabled, &twoFactorSecret, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert nullable fields
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if twoFactorSecret.Valid {
		user.TwoFactorSecret = twoFactorSecret.String
	}

	return &user, nil
}

// UpdateUser updates a user in the database
func (this *UserSqlDriver) UpdateUser(user *models.User) error {
	var database = this.db
	user.UpdatedAt = time.Now()

	_, err := database.Exec(models.UpdateUserSQL,
		user.Email, user.EmailVerified, user.EmailVerifiedAt, user.PasswordHash,
		user.FirstName, user.LastName, user.DisplayName, user.Avatar,
		user.Roles, user.IsActive, user.IsLocked, user.LockedUntil,
		user.FailedLoginCount, user.LastLoginAt, user.TwoFactorEnabled,
		user.TwoFactorSecret, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// UpdatePassword updates a user's password in the database
func (this *UserSqlDriver) UpdatePassword(userID, passwordHash string) error {
	var database = this.db
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := database.Exec(query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}

// UpdateUserLoginAttempt updates user login attempt information
func (this *UserSqlDriver) UpdateUserLoginAttempt(userID string, failedCount int, lockedUntil *time.Time, lastLoginAt *time.Time) error {
	var database = this.db
	_, err := database.Exec(models.UpdateUserLoginAttemptSQL,
		failedCount, lockedUntil, lastLoginAt, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user login attempt: %w", err)
	}
	return nil
}

// DeleteUser deletes a user from the database
func (this *UserSqlDriver) DeleteUser(id string) error {
	var database = this.db

	_, err := database.Exec(models.DeleteUserSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves a list of users with pagination
func (this *UserSqlDriver) ListUsers(limit, offset int) ([]*models.User, error) {
	var database = this.db

	rows, err := database.Query(models.ListUsersSQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.EmailVerified, &user.EmailVerifiedAt,
			&user.FirstName, &user.LastName, &user.DisplayName, &user.Avatar,
			&user.Roles, &user.IsActive, &user.IsLocked, &user.LockedUntil,
			&user.FailedLoginCount, &user.LastLoginAt, &user.TwoFactorEnabled,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// EmailExists checks if an email address is already registered
func (this *UserSqlDriver) EmailExists(email string) (bool, error) {
	var database = this.db

	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// IsUserLocked checks if a user account is currently locked
func (this *UserSqlDriver) IsUserLocked(user *models.User) bool {
	if !user.IsLocked {
		return false
	}

	if user.LockedUntil != nil && time.Now().After(*user.LockedUntil) {
		// Lock period has expired, unlock the user
		user.IsLocked = false
		user.LockedUntil = nil
		user.FailedLoginCount = 0
		this.UpdateUser(user)
		return false
	}

	return true
}
