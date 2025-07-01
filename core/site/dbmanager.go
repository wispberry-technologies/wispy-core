package site

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/core"
	"wispy-core/core/tenant/databases"

	_ "github.com/mattn/go-sqlite3"
)

// dbConnection represents a cached database connection
type dbConnection struct {
	db         *sql.DB
	lastUsed   time.Time
	usageCount int64
}

// databaseManager implements DatabaseManager interface
type databaseManager struct {
	mu          sync.RWMutex
	siteDomain  string
	dbDir       string
	connections map[string]*dbConnection
	maxIdle     int
	maxOpen     int
}

// NewDatabaseManager creates a new database manager for a site
func NewDatabaseManager(siteDomain string) core.DatabaseManager {
	globalConf := config.GlobalConf
	if globalConf == nil {
		common.Fatal("Global configuration not initialized")
	}

	// Construct database directory path: SitesPath + Domain + /databases/sqlite/
	dbDir := filepath.Join(
		globalConf.GetSitesPath(),
		siteDomain,
		"databases",
		"sqlite",
	)

	// Ensure database directory exists
	if err := common.EnsureDir(dbDir); err != nil {
		common.Error("Failed to create database directory %s: %v", dbDir, err)
	}

	return &databaseManager{
		siteDomain:  siteDomain,
		dbDir:       dbDir,
		connections: make(map[string]*dbConnection),
		maxIdle:     5,  // Maximum idle connections per database
		maxOpen:     25, // Maximum open connections per database
	}
}

// GetConnection returns a cached or new database connection
func (dm *databaseManager) GetConnection(dbName string) (*sql.DB, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Validate database name to prevent directory traversal
	if err := dm.validateDatabaseName(dbName); err != nil {
		return nil, fmt.Errorf("invalid database name: %v", err)
	}

	// Check if database file exists
	dbPath := filepath.Join(dm.dbDir, dbName+".db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database %s does not exist", dbName)
	}

	return dm.getConnectionInternal(dbName)
}

// GetOrCreateConnection gets a connection or creates the database if it doesn't exist
func (dm *databaseManager) GetOrCreateConnection(dbName string) (*sql.DB, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Validate database name to prevent directory traversal
	if err := dm.validateDatabaseName(dbName); err != nil {
		return nil, fmt.Errorf("invalid database name: %v", err)
	}

	// Check if database file exists
	dbPath := filepath.Join(dm.dbDir, dbName+".db")
	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
	}

	// If database doesn't exist, create it
	if !dbExists {
		if err := dm.createDatabaseWithScaffolding(dbName); err != nil {
			return nil, fmt.Errorf("failed to create database %s: %v", dbName, err)
		}
		common.Info("Created new database: %s", dbPath)
	}

	return dm.getConnectionInternal(dbName)
}

// getConnectionInternal handles the actual connection logic (assumes mutex is already held)
func (dm *databaseManager) getConnectionInternal(dbName string) (*sql.DB, error) {
	// Check if we have a cached connection
	if conn, exists := dm.connections[dbName]; exists {
		// Update usage statistics
		conn.lastUsed = time.Now()
		conn.usageCount++

		// Test connection health
		if err := conn.db.Ping(); err == nil {
			return conn.db, nil
		}

		// Connection is stale, remove it
		common.Warning("Database connection for %s is stale, recreating", dbName)
		conn.db.Close()
		delete(dm.connections, dbName)
	}

	// Create new connection
	dbPath := filepath.Join(dm.dbDir, dbName+".db")
	db, err := dm.createConnection(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %v", err)
	}

	// Cache the connection
	dm.connections[dbName] = &dbConnection{
		db:         db,
		lastUsed:   time.Now(),
		usageCount: 1,
	}

	common.Debug("Created new database connection for %s at %s", dbName, dbPath)
	return db, nil
}

// createDatabaseWithScaffolding creates a new database and runs its scaffolding function
func (dm *databaseManager) createDatabaseWithScaffolding(dbName string) error {
	// Check if we have a scaffolding function for this database
	scaffoldFunc, exists := databases.GetDatabaseScaffoldFunc(dbName)
	if !exists {
		return fmt.Errorf("no scaffolding function found for database '%s'", dbName)
	}

	// Create the database file by opening a connection
	dbPath := filepath.Join(dm.dbDir, dbName+".db")
	db, err := dm.createConnection(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %v", err)
	}
	defer db.Close()

	// Run the scaffolding function
	if err := scaffoldFunc(db); err != nil {
		// If scaffolding fails, remove the database file
		os.Remove(dbPath)
		return fmt.Errorf("failed to scaffold database: %v", err)
	}

	return nil
}

// createConnection creates a new SQLite database connection with proper settings
func (dm *databaseManager) createConnection(dbPath string) (*sql.DB, error) {
	// SQLite connection string with optimizations
	connStr := fmt.Sprintf("%s?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON", dbPath)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxIdleConns(dm.maxIdle)
	db.SetMaxOpenConns(dm.maxOpen)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 15)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// CreateDatabase creates a new database file if it doesn't exist
func (dm *databaseManager) CreateDatabase(dbName string) error {
	if err := dm.validateDatabaseName(dbName); err != nil {
		return fmt.Errorf("invalid database name: %v", err)
	}

	dbPath := filepath.Join(dm.dbDir, dbName+".db")

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database %s already exists", dbName)
	}

	// Create the database using GetOrCreateConnection
	_, err := dm.GetOrCreateConnection(dbName)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %v", dbName, err)
	}

	return nil
}

// ListDatabases returns all available databases for the site
func (dm *databaseManager) ListDatabases() ([]string, error) {
	files, err := os.ReadDir(dm.dbDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read database directory: %v", err)
	}

	var databases []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".db" {
			// Remove .db extension
			dbName := file.Name()[:len(file.Name())-3]
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// ExecuteSchema executes a schema file on the specified database
func (dm *databaseManager) ExecuteSchema(dbName, schemaPath string) error {
	if err := dm.validateDatabaseName(dbName); err != nil {
		return fmt.Errorf("invalid database name: %v", err)
	}

	// Read schema file
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Get database connection
	db, err := dm.GetConnection(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %v", err)
	}

	// Execute schema
	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	common.Info("Executed schema %s on database %s", schemaPath, dbName)
	return nil
}

// Close closes all database connections
func (dm *databaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var errors []error
	for dbName, conn := range dm.connections {
		if err := conn.db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database %s: %v", dbName, err))
		}
	}

	// Clear connections map
	dm.connections = make(map[string]*dbConnection)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// validateDatabaseName ensures the database name is safe and doesn't contain path traversal
func (dm *databaseManager) validateDatabaseName(dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	// Check for path traversal attempts
	if filepath.Clean(dbName) != dbName {
		return fmt.Errorf("database name contains invalid characters")
	}

	// Check for directory separators
	if filepath.Dir(dbName) != "." {
		return fmt.Errorf("database name cannot contain path separators")
	}

	// Check for reserved names
	reserved := []string{".", "..", "con", "prn", "aux", "nul"}
	for _, r := range reserved {
		if dbName == r {
			return fmt.Errorf("database name '%s' is reserved", dbName)
		}
	}

	return nil
}

// CleanupStaleConnections removes connections that haven't been used recently
func (dm *databaseManager) CleanupStaleConnections() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour) // Remove connections older than 1 hour

	for dbName, conn := range dm.connections {
		if conn.lastUsed.Before(cutoff) {
			common.Debug("Cleaning up stale connection for database %s", dbName)
			conn.db.Close()
			delete(dm.connections, dbName)
		}
	}
}

// GetConnectionStats returns statistics about database connections
func (dm *databaseManager) GetConnectionStats() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_connections"] = len(dm.connections)
	stats["site_domain"] = dm.siteDomain
	stats["database_directory"] = dm.dbDir

	connections := make(map[string]interface{})
	for dbName, conn := range dm.connections {
		connections[dbName] = map[string]interface{}{
			"last_used":   conn.lastUsed,
			"usage_count": conn.usageCount,
		}
	}
	stats["connections"] = connections

	return stats
}
