package databases

import (
	"database/sql"
	"fmt"
	"wispy-core/common"
)

// ScaffoldContentDatabase creates the schema for the content database
func ScaffoldContentDatabase(db *sql.DB) error {
	common.Info("Scaffolding content database")

	// Create content table
	contentTableSQL := `
    CREATE TABLE IF NOT EXISTS content (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        content_type TEXT DEFAULT 'page',
        status TEXT DEFAULT 'draft',
        author_id INTEGER,
        published_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create content meta table
	contentMetaTableSQL := `
    CREATE TABLE IF NOT EXISTS content_meta (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content_id INTEGER NOT NULL,
        meta_key TEXT NOT NULL,
        meta_value TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (content_id) REFERENCES content(id) ON DELETE CASCADE,
        UNIQUE(content_id, meta_key)
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_content_slug ON content(slug);`,
		`CREATE INDEX IF NOT EXISTS idx_content_type ON content(content_type);`,
		`CREATE INDEX IF NOT EXISTS idx_content_status ON content(status);`,
		`CREATE INDEX IF NOT EXISTS idx_content_published_at ON content(published_at);`,
		`CREATE INDEX IF NOT EXISTS idx_content_meta_key ON content_meta(meta_key);`,
		`CREATE INDEX IF NOT EXISTS idx_content_meta_content_id ON content_meta(content_id);`,
	}

	// Execute table creation
	if _, err := db.Exec(contentTableSQL); err != nil {
		return fmt.Errorf("failed to create content table: %v", err)
	}

	if _, err := db.Exec(contentMetaTableSQL); err != nil {
		return fmt.Errorf("failed to create content_meta table: %v", err)
	}

	// Execute indexes
	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
