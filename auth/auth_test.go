package auth

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	email := "test@example.com"
	firstName := "John"
	lastName := "Doe"
	displayName := "John Doe"

	user := NewUser(email, firstName, lastName, displayName)

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.FirstName != firstName {
		t.Errorf("Expected first name %s, got %s", firstName, user.FirstName)
	}
	if user.LastName != lastName {
		t.Errorf("Expected last name %s, got %s", lastName, user.LastName)
	}
	if user.DisplayName != displayName {
		t.Errorf("Expected display name %s, got %s", displayName, user.DisplayName)
	}
	if user.ID == "" {
		t.Error("Expected user ID to be generated")
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestNewSession(t *testing.T) {
	userID := "test-user-id"
	duration := time.Hour
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"

	session := NewSession(userID, duration, ipAddress, userAgent)

	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}
	if session.IPAddress != ipAddress {
		t.Errorf("Expected IP address %s, got %s", ipAddress, session.IPAddress)
	}
	if session.UserAgent != userAgent {
		t.Errorf("Expected user agent %s, got %s", userAgent, session.UserAgent)
	}
	if session.Token == "" {
		t.Error("Expected session token to be generated")
	}
	if session.ID == "" {
		t.Error("Expected session ID to be generated")
	}
}

func TestUserRoles(t *testing.T) {
	user := NewUser("test@example.com", "John", "Doe", "John Doe")

	// Test initial empty roles
	roles := user.GetRoles()
	if len(roles) != 0 {
		t.Errorf("Expected empty roles, got %v", roles)
	}

	// Test setting roles
	testRoles := []string{"admin", "user"}
	user.SetRoles(testRoles)

	roles = user.GetRoles()
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
	if roles[0] != "admin" || roles[1] != "user" {
		t.Errorf("Expected [admin, user], got %v", roles)
	}

	// Test HasRole
	if !user.HasRole("admin") {
		t.Error("Expected user to have admin role")
	}
	if !user.HasRole("user") {
		t.Error("Expected user to have user role")
	}
	if user.HasRole("moderator") {
		t.Error("Expected user to not have moderator role")
	}
}

func TestSessionExpiration(t *testing.T) {
	userID := "test-user-id"
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"

	// Create session with very short duration
	session := NewSession(userID, time.Nanosecond, ipAddress, userAgent)

	// Wait a bit to ensure expiration
	time.Sleep(time.Millisecond)

	if !session.IsExpired() {
		t.Error("Expected session to be expired")
	}

	if session.IsValid() {
		t.Error("Expected session to be invalid")
	}
}
