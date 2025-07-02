package databases

import (
	"database/sql"
	"wispy-core/common"
)

// DatabaseScaffoldFunc represents a function that scaffolds a database schema
type DatabaseScaffoldFunc func(db *sql.DB) error

// DatabaseManager handles database connections for a site
type Manager interface {
	// GetConnection returns a database connection for the specified database name
	GetConnection(dbName string) (*sql.DB, error)
	// CreateDatabase creates a new database file if it doesn't exist
	CreateDatabase(dbName string) error
	// ListDatabases returns all available databases for the site
	ListDatabases() ([]string, error)
	// Close closes all database connections
	Close() error
	// ExecuteSchema executes a schema file on the specified database
	ExecuteSchema(dbName, schemaPath string) error
	// GetOrCreateConnection returns a database connection, creating it if it doesn't exist
	GetOrCreateConnection(dbName string) (*sql.DB, error)
}

// DatabaseScaffolds contains the mapping of database names to their scaffolding functions
var DatabaseScaffolds = map[string]DatabaseScaffoldFunc{
	"forms":     ScaffoldFormsDatabase,
	"users":     ScaffoldUsersDatabase,
	"analytics": ScaffoldAnalyticsDatabase,
	"content":   ScaffoldContentDatabase,
	"media":     ScaffoldMediaDatabase,
}

// GetDatabaseScaffoldFunc returns the scaffolding function for a given database name
func GetDatabaseScaffoldFunc(dbName string) (DatabaseScaffoldFunc, bool) {
	scaffoldFunc, exists := DatabaseScaffolds[dbName]
	return scaffoldFunc, exists
}

// ListAvailableDatabases returns all database names that can be scaffolded
func ListAvailableDatabases() []string {
	var databases []string
	for dbName := range DatabaseScaffolds {
		databases = append(databases, dbName)
	}
	return databases
}

// RegisterDatabaseScaffold allows registering new database scaffolding functions
func RegisterDatabaseScaffold(dbName string, scaffoldFunc DatabaseScaffoldFunc) {
	DatabaseScaffolds[dbName] = scaffoldFunc
	common.Info("Registered database scaffold for: %s", dbName)
}
