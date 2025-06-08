package common

import (
	"fmt"

	"wispy-core/models"
)

// Authentication methods for SiteInstanceManager
// createAuthTables creates all necessary authentication tables for a site
func CreateAuthTables(instance *SiteInstance) error {
	db, err := instance.GetDB("users")
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create users table
	if _, err := db.Exec(models.CreateUserTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sessions table
	if _, err := db.Exec(models.CreateSessionTableSQL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create OAuth accounts table
	if _, err := db.Exec(models.CreateOAuthAccountTableSQL); err != nil {
		return fmt.Errorf("failed to create oauth accounts table: %w", err)
	}

	return nil
}
