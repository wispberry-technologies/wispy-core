package auth

// SQL queries for authentication operations

// User table queries
const (
	CreateUserTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			email_verified BOOLEAN DEFAULT FALSE,
			email_verified_at DATETIME,
			password_hash TEXT,
			first_name TEXT,
			last_name TEXT,
			display_name TEXT,
			avatar TEXT,
			roles TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			is_locked BOOLEAN DEFAULT FALSE,
			locked_until DATETIME,
			failed_login_count INTEGER DEFAULT 0,
			last_login_at DATETIME,
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			two_factor_secret TEXT,
			oauth_providers TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
	`

	InsertUserSQL = `
		INSERT INTO users (
			id, email, email_verified, password_hash, first_name, last_name, 
			display_name, avatar, roles, is_active, is_locked, failed_login_count, 
			two_factor_enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	GetUserByEmailSQL = `
		SELECT id, email, email_verified, email_verified_at, password_hash, 
			   first_name, last_name, display_name, avatar, roles, is_active, 
			   is_locked, locked_until, failed_login_count, last_login_at, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at
		FROM users 
		WHERE email = ? AND is_active = true
	`

	GetUserByIDSQL = `
		SELECT id, email, email_verified, email_verified_at, password_hash, 
			   first_name, last_name, display_name, avatar, roles, is_active, 
			   is_locked, locked_until, failed_login_count, last_login_at, 
			   two_factor_enabled, two_factor_secret, created_at, updated_at
		FROM users 
		WHERE id = ?
	`

	UpdateUserSQL = `
		UPDATE users 
		SET email = ?, email_verified = ?, email_verified_at = ?, password_hash = ?, 
			first_name = ?, last_name = ?, display_name = ?, avatar = ?, roles = ?, 
			is_active = ?, is_locked = ?, locked_until = ?, failed_login_count = ?, 
			last_login_at = ?, two_factor_enabled = ?, two_factor_secret = ?, 
			updated_at = ?
		WHERE id = ?
	`

	UpdateUserLoginAttemptSQL = `
		UPDATE users 
		SET failed_login_count = ?, locked_until = ?, last_login_at = ?, updated_at = ?
		WHERE id = ?
	`

	DeleteUserSQL = `
		DELETE FROM users WHERE id = ?
	`

	ListUsersSQL = `
		SELECT id, email, email_verified, email_verified_at, first_name, last_name, 
			   display_name, avatar, roles, is_active, is_locked, locked_until, 
			   failed_login_count, last_login_at, two_factor_enabled, created_at, updated_at
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
)

// Session table queries
const (
	CreateSessionTableSQL = `
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	InsertSessionSQL = `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at, updated_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	GetSessionByTokenSQL = `
		SELECT id, user_id, token, expires_at, created_at, updated_at, ip_address, user_agent
		FROM sessions 
		WHERE token = ? AND expires_at > datetime('now')
	`

	GetSessionsByUserIDSQL = `
		SELECT id, user_id, token, expires_at, created_at, updated_at, ip_address, user_agent
		FROM sessions 
		WHERE user_id = ? AND expires_at > datetime('now')
		ORDER BY created_at DESC
	`

	UpdateSessionSQL = `
		UPDATE sessions 
		SET expires_at = ?, updated_at = ?
		WHERE id = ?
	`

	DeleteSessionSQL = `
		DELETE FROM sessions WHERE id = ?
	`

	DeleteSessionByTokenSQL = `
		DELETE FROM sessions WHERE token = ?
	`

	DeleteExpiredSessionsSQL = `
		DELETE FROM sessions WHERE expires_at <= datetime('now')
	`

	DeleteUserSessionsSQL = `
		DELETE FROM sessions WHERE user_id = ?
	`
)

// OAuth account queries
const (
	CreateOAuthAccountTableSQL = `
		CREATE TABLE IF NOT EXISTS oauth_accounts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			email TEXT,
			display_name TEXT,
			avatar TEXT,
			access_token TEXT,
			refresh_token TEXT,
			expires_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(provider, provider_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);
		CREATE INDEX IF NOT EXISTS idx_oauth_accounts_provider ON oauth_accounts(provider, provider_id);
	`

	InsertOAuthAccountSQL = `
		INSERT INTO oauth_accounts (
			id, user_id, provider, provider_id, email, display_name, avatar, 
			access_token, refresh_token, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	GetOAuthAccountSQL = `
		SELECT id, user_id, provider, provider_id, email, display_name, avatar, 
			   access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts 
		WHERE provider = ? AND provider_id = ?
	`

	GetOAuthAccountsByUserIDSQL = `
		SELECT id, user_id, provider, provider_id, email, display_name, avatar, 
			   access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts 
		WHERE user_id = ?
	`

	UpdateOAuthAccountSQL = `
		UPDATE oauth_accounts 
		SET email = ?, display_name = ?, avatar = ?, access_token = ?, 
			refresh_token = ?, expires_at = ?, updated_at = ?
		WHERE id = ?
	`

	DeleteOAuthAccountSQL = `
		DELETE FROM oauth_accounts WHERE id = ?
	`
)
