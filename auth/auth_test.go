package auth

import (
	_ "github.com/mattn/go-sqlite3"
)

// // Test helpers and setup functions

// // setupTestDB creates an in-memory SQLite database for testing
// func setupTestDB(t *testing.T) *sql.DB {
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	if err != nil {
// 		t.Fatalf("Failed to open in-memory database: %v", err)
// 	}

// 	// Create the necessary tables
// 	_, err = db.Exec(`
// 		CREATE TABLE IF NOT EXISTS users (
// 			id TEXT PRIMARY KEY,
// 			email TEXT UNIQUE NOT NULL,
// 			username TEXT UNIQUE NOT NULL,
// 			password TEXT NOT NULL,
// 			display_name TEXT,
// 			roles TEXT,
// 			metadata BLOB,
// 			created_at DATETIME NOT NULL,
// 			updated_at DATETIME NOT NULL,
// 			last_login DATETIME,
// 			oauth_provider TEXT,
// 			oauth_id TEXT
// 		);

// 		CREATE TABLE IF NOT EXISTS sessions (
// 			id TEXT PRIMARY KEY,
// 			user_id TEXT NOT NULL,
// 			token TEXT UNIQUE NOT NULL,
// 			expires_at DATETIME NOT NULL,
// 			ip TEXT,
// 			user_agent TEXT,
// 			created_at DATETIME NOT NULL,
// 			updated_at DATETIME NOT NULL,
// 			data BLOB,
// 			FOREIGN KEY (user_id) REFERENCES users(id)
// 		);
// 	`)
// 	if err != nil {
// 		t.Fatalf("Failed to create tables: %v", err)
// 	}

// 	return db
// }

// // createTestUser creates a test user in the database and returns the user object
// func createTestUser(t *testing.T, userStore UserStore) *User {
// 	user := &User{
// 		Email:    "test@example.com",
// 		Username: "testuser",
// 		Password: "password123", // Will be hashed by the store
// 		Roles:    []string{"user"},
// 	}

// 	err := userStore.CreateUser(context.Background(), user)
// 	if err != nil {
// 		t.Fatalf("Failed to create test user: %v", err)
// 	}
// 	return user
// }

// // setupTestAuth creates a test auth provider with in-memory database
// func setupTestAuth(t *testing.T) (*defaultAuthProvider, *sql.DB) {
// 	// Create in-memory test database
// 	db := setupTestDB(t)

// 	// Create stores with our test DB connection
// 	userStore, err := NewSQLiteUserStore(db)
// 	if err != nil {
// 		t.Fatalf("Failed to create user store: %v", err)
// 	}

// 	sessionStore, err := NewSQLiteSessionStore(db)
// 	if err != nil {
// 		t.Fatalf("Failed to create session store: %v", err)
// 	}

// 	// Create a config for testing
// 	config := Config{
// 		DBType:           "sqlite",
// 		DBConn:           ":memory:",
// 		TokenSecret:      "test-secret-key",
// 		TokenExpiration:  time.Hour,
// 		PasswordMinChars: 6,
// 		AllowSignup:      true,
// 		CookieName:       "test_auth_cookie",
// 		CookieSecure:     false,
// 		CookieHTTPOnly:   true,
// 	}

// 	// Create a proper DefaultAuthProvider for testing
// 	config.DBConn = ":memory:" // Make sure we use in-memory
// 	authProvider, err := NewDefaultAuthProvider(config)
// 	if err != nil {
// 		t.Fatalf("Failed to create auth provider: %v", err)
// 	}

// 	// Replace the stores with our test stores
// 	authProvider.userStore = userStore
// 	authProvider.sessionStore = sessionStore

// 	return authProvider, db
// }

// func TestMain(m *testing.M) {
// 	// Run tests
// 	code := m.Run()

// 	// Exit
// 	os.Exit(code)
// }
