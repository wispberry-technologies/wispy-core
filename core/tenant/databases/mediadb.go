package databases

import (
	"database/sql"
	"fmt"
	"wispy-core/common"
)

// ScaffoldMediaDatabase creates the schema for the media database
func ScaffoldMediaDatabase(db *sql.DB) error {
	common.Info("Scaffolding media database")

	// Create media table
	mediaTableSQL := `
    CREATE TABLE IF NOT EXISTS media (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT NOT NULL UNIQUE,
        filename TEXT NOT NULL,
        original_filename TEXT NOT NULL,
        file_path TEXT NOT NULL,
        file_size INTEGER NOT NULL,
        mime_type TEXT NOT NULL,
        width INTEGER,
        height INTEGER,
        alt_text TEXT,
        title TEXT,
        description TEXT,
        uploaded_by INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create media metadata table
	mediaMetaTableSQL := `
    CREATE TABLE IF NOT EXISTS media_metadata (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        media_id INTEGER NOT NULL,
        meta_key TEXT NOT NULL,
        meta_value TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE,
        UNIQUE(media_id, meta_key)
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_media_uuid ON media(uuid);`,
		`CREATE INDEX IF NOT EXISTS idx_media_filename ON media(filename);`,
		`CREATE INDEX IF NOT EXISTS idx_media_mime_type ON media(mime_type);`,
		`CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_media_uploaded_by ON media(uploaded_by);`,
		`CREATE INDEX IF NOT EXISTS idx_media_meta_key ON media_metadata(meta_key);`,
		`CREATE INDEX IF NOT EXISTS idx_media_meta_media_id ON media_metadata(media_id);`,
	}

	// Execute table creation
	if _, err := db.Exec(mediaTableSQL); err != nil {
		return fmt.Errorf("failed to create media table: %v", err)
	}

	if _, err := db.Exec(mediaMetaTableSQL); err != nil {
		return fmt.Errorf("failed to create media_metadata table: %v", err)
	}

	// Execute indexes
	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
