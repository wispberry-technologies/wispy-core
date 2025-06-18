package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// IsValidPassword checks if a password meets minimum requirements
func IsValidPassword(password string) bool {
	// Basic password validation - minimum 8 characters
	// In a real implementation, you might want more sophisticated rules
	return len(password) >= 8
}

// Helper functions for handling API responses
// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// ToUserInfo converts a User to UserInfo for API responses (removes sensitive data)
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and not empty)
func (s *Session) IsValid() bool {
	return s.Token != "" && !s.IsExpired()
}

// generateID generates a random ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken generates a random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
