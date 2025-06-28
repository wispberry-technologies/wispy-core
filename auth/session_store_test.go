package auth

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSessionStore_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userStore, err := NewSQLiteUserStore(db)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	sessionStore, err := NewSQLiteSessionStore(db)
	if err != nil {
		t.Fatalf("Failed to create session store: %v", err)
	}

	// Create a test user
	user := &User{
		Email:    "session@example.com",
		Username: "sessionuser",
		Password: "password123",
	}

	err = userStore.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a session
	session := &Session{
		UserID:    user.ID,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(time.Hour),
		IP:        "127.0.0.1",
		UserAgent: "Test User Agent",
	}

	err = sessionStore.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Check that the session has an ID
	if session.ID == "" {
		t.Error("Session ID not generated")
	}

	// Get the session by token
	fetchedSession, err := sessionStore.GetSessionByToken(context.Background(), session.Token)
	if err != nil {
		t.Fatalf("Failed to get session by token: %v", err)
	}

	// Check that the session properties match
	if fetchedSession.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, fetchedSession.UserID)
	}
	if fetchedSession.Token != "test-token" {
		t.Errorf("Expected Token %s, got %s", "test-token", fetchedSession.Token)
	}
	if fetchedSession.IP != "127.0.0.1" {
		t.Errorf("Expected IP %s, got %s", "127.0.0.1", fetchedSession.IP)
	}
}

func TestSessionStore_DeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userStore, err := NewSQLiteUserStore(db)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	sessionStore, err := NewSQLiteSessionStore(db)
	if err != nil {
		t.Fatalf("Failed to create session store: %v", err)
	}

	// Create a test user
	user := &User{
		Email:    "sessiondelete@example.com",
		Username: "sessiondeleteuser",
		Password: "password123",
	}

	err = userStore.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a session
	session := &Session{
		UserID:    user.ID,
		Token:     "delete-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err = sessionStore.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// First, get the session to get its ID
	fetchedSession, err := sessionStore.GetSessionByToken(context.Background(), session.Token)
	if err != nil {
		t.Fatalf("Failed to get session by token: %v", err)
	}

	// Delete the session by ID
	err = sessionStore.DeleteSession(context.Background(), fetchedSession.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Try to get the deleted session
	_, err = sessionStore.GetSessionByToken(context.Background(), session.Token)
	if err == nil {
		t.Error("Expected error when getting deleted session, but got nil")
	}
}

func TestSessionStore_ListSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userStore, err := NewSQLiteUserStore(db)
	if err != nil {
		t.Fatalf("Failed to create user store: %v", err)
	}

	sessionStore, err := NewSQLiteSessionStore(db)
	if err != nil {
		t.Fatalf("Failed to create session store: %v", err)
	}

	// Create a test user
	user := &User{
		Email:    "listsessions@example.com",
		Username: "listsessionsuser",
		Password: "password123",
	}

	err = userStore.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &Session{
			UserID:    user.ID,
			Token:     fmt.Sprintf("token-%d", i),
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err = sessionStore.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
	}

	// List sessions for the user
	sessions, err := sessionStore.GetUserSessions(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	// Check that we have 3 sessions
	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}
