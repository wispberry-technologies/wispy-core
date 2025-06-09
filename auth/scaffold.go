package auth

import (
	"fmt"

	"wispy-core/cache"
	"wispy-core/models"
)

// Authentication methods for SiteInstanceManager
// createAuthTables creates all necessary authentication tables for a site
func CreateAuthTables(instance *models.SiteInstance) error {
	db, err := cache.GetDB(instance, "users")
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
