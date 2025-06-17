package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

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

// TODO: Reevaluate the need for this
// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func SendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	response := &APIResponse{
		Success: true,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func SendError(w http.ResponseWriter, statusCode int, message, details string) {
	response := &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    statusCode,
			Message: message,
			Details: details,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
