package databases

import (
	"database/sql"
	"fmt"
	"wispy-core/common"
)

// ScaffoldFormsDatabase creates the schema for the forms database
func ScaffoldFormsDatabase(db *sql.DB) error {
	common.Info("Scaffolding forms database")

	// Create forms table
	formsTableSQL := `
    CREATE TABLE IF NOT EXISTS forms (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        description TEXT,
        fields TEXT NOT NULL, -- JSON string containing form fields
        settings TEXT, -- JSON string containing form settings
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create form submissions table
	submissionsTableSQL := `
    CREATE TABLE IF NOT EXISTS form_submissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        form_id INTEGER NOT NULL,
        data TEXT NOT NULL, -- JSON string containing submission data
        ip_address TEXT,
        user_agent TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_forms_name ON forms(name);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_form_id ON form_submissions(form_id);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON form_submissions(created_at);`,
	}

	// Execute table creation
	if _, err := db.Exec(formsTableSQL); err != nil {
		return fmt.Errorf("failed to create forms table: %v", err)
	}

	if _, err := db.Exec(submissionsTableSQL); err != nil {
		return fmt.Errorf("failed to create form_submissions table: %v", err)
	}

	// Execute indexes
	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
