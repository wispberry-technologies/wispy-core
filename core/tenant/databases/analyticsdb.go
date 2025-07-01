package databases

import (
	"database/sql"
	"fmt"
	"wispy-core/common"
)

// ScaffoldAnalyticsDatabase creates the schema for the analytics database
func ScaffoldAnalyticsDatabase(db *sql.DB) error {
	common.Info("Scaffolding analytics database")

	// Create page views table
	pageViewsTableSQL := `
    CREATE TABLE IF NOT EXISTS page_views (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        page_path TEXT NOT NULL,
        page_title TEXT,
        referrer TEXT,
        user_agent TEXT,
        ip_address TEXT,
        session_id TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create events table
	eventsTableSQL := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_name TEXT NOT NULL,
        event_data TEXT, -- JSON string containing event data
        page_path TEXT,
        session_id TEXT,
        ip_address TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_page_views_path ON page_views(page_path);`,
		`CREATE INDEX IF NOT EXISTS idx_page_views_created_at ON page_views(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_page_views_session_id ON page_views(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_events_name ON events(event_name);`,
		`CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);`,
	}

	// Execute table creation
	if _, err := db.Exec(pageViewsTableSQL); err != nil {
		return fmt.Errorf("failed to create page_views table: %v", err)
	}

	if _, err := db.Exec(eventsTableSQL); err != nil {
		return fmt.Errorf("failed to create events table: %v", err)
	}

	// Execute indexes
	for _, indexSQL := range indexesSQL {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
