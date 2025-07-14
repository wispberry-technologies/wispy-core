package auth

// func TestUserStore_CreateAndGet(t *testing.T) {
// 	db := setupTestDB(t)
// 	defer db.Close()

// 	userStore, err := NewSQLiteUserStore(db)
// 	if err != nil {
// 		t.Fatalf("Failed to create user store: %v", err)
// 	}

// 	// Test user creation
// 	user := &User{
// 		Email:    "test@example.com",
// 		Username: "testuser",
// 		Password: "password123",
// 		Roles:    []string{"user"},
// 	}

// 	err = userStore.CreateUser(context.Background(), user)
// 	if err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}

// 	// Check that the user has an ID and the password is hashed
// 	if user.ID == "" {
// 		t.Error("User ID not generated")
// 	}
// 	if user.Password == "password123" {
// 		t.Error("Password was not hashed")
// 	}
// 	if user.CreatedAt.IsZero() {
// 		t.Error("CreatedAt timestamp not set")
// 	}
// 	if user.UpdatedAt.IsZero() {
// 		t.Error("UpdatedAt timestamp not set")
// 	}

// 	// Test getting user by ID
// 	fetchedUser, err := userStore.GetUserByID(context.Background(), user.ID)
// 	if err != nil {
// 		t.Fatalf("Failed to get user by ID: %v", err)
// 	}

// 	if fetchedUser.Email != user.Email {
// 		t.Errorf("Expected email %s, got %s", user.Email, fetchedUser.Email)
// 	}
// 	if fetchedUser.Username != user.Username {
// 		t.Errorf("Expected username %s, got %s", user.Username, fetchedUser.Username)
// 	}

// 	// Test getting user by email
// 	fetchedByEmail, err := userStore.GetUserByEmail(context.Background(), user.Email)
// 	if err != nil {
// 		t.Fatalf("Failed to get user by email: %v", err)
// 	}
// 	if fetchedByEmail.ID != user.ID {
// 		t.Errorf("Expected ID %s, got %s", user.ID, fetchedByEmail.ID)
// 	}

// 	// Test getting user by username
// 	fetchedByUsername, err := userStore.GetUserByUsername(context.Background(), user.Username)
// 	if err != nil {
// 		t.Fatalf("Failed to get user by username: %v", err)
// 	}
// 	if fetchedByUsername.ID != user.ID {
// 		t.Errorf("Expected ID %s, got %s", user.ID, fetchedByUsername.ID)
// 	}

// 	// Test getting a non-existent user
// 	_, err = userStore.GetUserByEmail(context.Background(), "nonexistent@example.com")
// 	if err == nil {
// 		t.Error("Expected error when getting non-existent user, but got nil")
// 	}
// }

// func TestUserStore_UpdateUser(t *testing.T) {
// 	db := setupTestDB(t)
// 	defer db.Close()

// 	userStore, err := NewSQLiteUserStore(db)
// 	if err != nil {
// 		t.Fatalf("Failed to create user store: %v", err)
// 	}

// 	// Create a test user
// 	user := &User{
// 		Email:    "update@example.com",
// 		Username: "updateuser",
// 		Password: "password123",
// 		Roles:    []string{"user"},
// 	}

// 	err = userStore.CreateUser(context.Background(), user)
// 	if err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}

// 	// Update user properties
// 	user.DisplayName = "Test User"
// 	user.Roles = []string{"user", "admin"}

// 	err = userStore.UpdateUser(context.Background(), user)
// 	if err != nil {
// 		t.Fatalf("Failed to update user: %v", err)
// 	}

// 	// Fetch the updated user
// 	updated, err := userStore.GetUserByID(context.Background(), user.ID)
// 	if err != nil {
// 		t.Fatalf("Failed to get updated user: %v", err)
// 	}

// 	// Check that the properties were updated
// 	if updated.DisplayName != "Test User" {
// 		t.Errorf("Expected DisplayName 'Test User', got %s", updated.DisplayName)
// 	}

// 	// Check roles in an order-insensitive way
// 	if len(updated.Roles) != 2 {
// 		t.Errorf("Expected 2 roles, got %d: %v", len(updated.Roles), updated.Roles)
// 	} else {
// 		// Create a map to check role existence regardless of order
// 		roleMap := make(map[string]bool)
// 		for _, role := range updated.Roles {
// 			roleMap[role] = true
// 		}

// 		if !roleMap["user"] {
// 			t.Errorf("Expected role 'user' not found in roles: %v", updated.Roles)
// 		}
// 		if !roleMap["admin"] {
// 			t.Errorf("Expected role 'admin' not found in roles: %v", updated.Roles)
// 		}
// 	}
// }

// func TestUserStore_DeleteUser(t *testing.T) {
// 	db := setupTestDB(t)
// 	defer db.Close()

// 	userStore, err := NewSQLiteUserStore(db)
// 	if err != nil {
// 		t.Fatalf("Failed to create user store: %v", err)
// 	}

// 	// Create a test user
// 	user := &User{
// 		Email:    "delete@example.com",
// 		Username: "deleteuser",
// 		Password: "password123",
// 	}

// 	err = userStore.CreateUser(context.Background(), user)
// 	if err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}

// 	// Delete the user
// 	err = userStore.DeleteUser(context.Background(), user.ID)
// 	if err != nil {
// 		t.Fatalf("Failed to delete user: %v", err)
// 	}

// 	// Try to fetch the deleted user
// 	_, err = userStore.GetUserByID(context.Background(), user.ID)
// 	if err == nil {
// 		t.Error("Expected error when getting deleted user, but got nil")
// 	}
// }
