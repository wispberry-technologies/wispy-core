package auth

// func TestAuthProvider_RegisterAndLogin(t *testing.T) {
// 	authProvider, db := setupTestAuth(t)
// 	defer db.Close()

// 	// Test registration
// 	testEmail := "auth@example.com"
// 	testUsername := "authuser"
// 	testPassword := "authpass123"

// 	user, err := authProvider.Register(context.Background(), testEmail, testUsername, testPassword)
// 	if err != nil {
// 		t.Fatalf("Failed to register user: %v", err)
// 	}

// 	if user.Email != testEmail {
// 		t.Errorf("Expected email %s, got %s", testEmail, user.Email)
// 	}
// 	if user.Username != testUsername {
// 		t.Errorf("Expected username %s, got %s", testUsername, user.Username)
// 	}

// 	// Test login
// 	session, err := authProvider.Login(context.Background(), testEmail, testPassword)
// 	if err != nil {
// 		t.Fatalf("Failed to login: %v", err)
// 	}

// 	if session.UserID != user.ID {
// 		t.Errorf("Session UserID %s doesn't match user ID %s", session.UserID, user.ID)
// 	}
// 	if session.Token == "" {
// 		t.Error("Session token is empty")
// 	}

// 	// Test login with incorrect password
// 	_, err = authProvider.Login(context.Background(), testEmail, "wrongpassword")
// 	if err == nil {
// 		t.Error("Expected login to fail with incorrect password, but it succeeded")
// 	}

// 	// Test login with incorrect email
// 	_, err = authProvider.Login(context.Background(), "wrong@example.com", testPassword)
// 	if err == nil {
// 		t.Error("Expected login to fail with incorrect email, but it succeeded")
// 	}
// }

// func TestAuthProvider_ValidateAndRefreshSession(t *testing.T) {
// 	authProvider, db := setupTestAuth(t)
// 	defer db.Close()

// 	// Register a test user
// 	user, err := authProvider.Register(context.Background(), "session@example.com", "sessionuser", "password123")
// 	if err != nil {
// 		t.Fatalf("Failed to register user: %v", err)
// 	}

// 	// Create a session
// 	session, err := authProvider.Login(context.Background(), "session@example.com", "password123")
// 	if err != nil {
// 		t.Fatalf("Failed to login: %v", err)
// 	}

// 	// Validate the session
// 	validatedSession, validatedUser, err := authProvider.ValidateSession(context.Background(), session.Token)
// 	if err != nil {
// 		t.Fatalf("Failed to validate session: %v", err)
// 	}

// 	if validatedSession.ID != session.ID {
// 		t.Errorf("Validated session ID %s doesn't match original %s", validatedSession.ID, session.ID)
// 	}
// 	if validatedUser.ID != user.ID {
// 		t.Errorf("Validated user ID %s doesn't match original %s", validatedUser.ID, user.ID)
// 	}

// 	// Store the original token for comparison
// 	originalToken := session.Token

// 	// Refresh the session
// 	newSession, err := authProvider.RefreshSession(context.Background(), session.Token)
// 	if err != nil {
// 		t.Fatalf("Failed to refresh session: %v", err)
// 	}

// 	// Print token details for debugging
// 	t.Logf("Original token: %s", originalToken[:10])
// 	t.Logf("New token: %s", newSession.Token[:10])

// 	// Verify token has changed
// 	if newSession.Token == originalToken {
// 		t.Error("Refreshed session token was not updated")
// 	}

// 	// Verify user ID is preserved
// 	if newSession.UserID != session.UserID {
// 		t.Errorf("Refreshed session user ID %s doesn't match original %s", newSession.UserID, session.UserID)
// 	}

// 	// Verify the original token no longer works
// 	_, err = authProvider.GetSessionStore().GetSessionByToken(context.Background(), originalToken)
// 	if err == nil {
// 		t.Error("Original token still valid after refresh")
// 	}

// 	// Verify the new token works
// 	verifySession, err := authProvider.GetSessionStore().GetSessionByToken(context.Background(), newSession.Token)
// 	if err != nil {
// 		t.Errorf("New token is not valid in the database: %v", err)
// 	} else if verifySession.ID != newSession.ID {
// 		t.Errorf("Session ID mismatch after refresh: %s vs %s", verifySession.ID, newSession.ID)
// 	}
// }

// func TestAuthProvider_LogoutAndInvalidate(t *testing.T) {
// 	authProvider, db := setupTestAuth(t)
// 	defer db.Close()

// 	// Register and login a test user
// 	_, err := authProvider.Register(context.Background(), "logout@example.com", "logoutuser", "password123")
// 	if err != nil {
// 		t.Fatalf("Failed to register user: %v", err)
// 	}

// 	session, err := authProvider.Login(context.Background(), "logout@example.com", "password123")
// 	if err != nil {
// 		t.Fatalf("Failed to login: %v", err)
// 	}

// 	// Logout
// 	err = authProvider.Logout(context.Background(), session.Token)
// 	if err != nil {
// 		t.Fatalf("Failed to logout: %v", err)
// 	}

// 	// Try to validate the session after logout
// 	_, _, err = authProvider.ValidateSession(context.Background(), session.Token)
// 	if err == nil {
// 		t.Error("Expected session validation to fail after logout, but it succeeded")
// 	}
// }
