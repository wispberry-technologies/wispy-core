package databases

import (
	"database/sql"
	"fmt"
	"wispy-core/common"
)

// ScaffoldUsersDatabase creates the schema for the users database
func ScaffoldUsersDatabase(db *sql.DB) error {
	common.Info("Scaffolding users database")

	// Create users table
	usersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT,
			role TEXT DEFAULT 'user',
			active BOOLEAN DEFAULT 1,
			email_verified BOOLEAN DEFAULT 0,
			must_change_password BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		);`

	// Create user sessions table
	sessionsTableSQL := `
    CREATE TABLE IF NOT EXISTS user_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			session_token TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);`,
		`CREATE INDEX IF NOT EXISTS idx_users_id ON users(id);`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(session_token);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON user_sessions(expires_at);`,
	}

	// Execute table creation
	if _, err := db.Exec(usersTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if _, err := db.Exec(sessionsTableSQL); err != nil {
		return fmt.Errorf("failed to create user_sessions table: %v", err)
	}

	// Execute indexes
	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
