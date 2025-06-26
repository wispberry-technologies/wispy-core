package scaffold

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"wispy-core/internal/models"
)

// models.Page
//
//	type Page struct {
//		Title         string            `json:"title"`
//		Description   string            `json:"description"`
//		Lang          string            `json:"lang"`
//		Slug          string            `json:"slug"`
//		Keywords      []string          `json:"keywords"`
//		Author        string            `json:"author"`
//		LayoutName    string            `json:"layout"`
//		IsDraft       bool              `json:"is_draft"`
//		IsStatic      bool              `json:"is_static"`
//		RequireAuth   bool              `json:"require_auth"`
//		RequiredRoles []string          `json:"required_roles"`
//		FilePath      string            `json:"file_path"`
//		Protected     string            `json:"protected"`
//		PageData      map[string]string `json:"custom_data"`
//		CreatedAt     time.Time         `json:"created_at"`
//		UpdatedAt     time.Time         `json:"updated_at"`
//		PublishedAt   *time.Time        `json:"published_at,omitempty"`
//		MetaTags      []HtmlMetaTag     `json:"meta_tags"`
//		// SiteDetails contains information about the site this page belongs to
//		SiteDetails SiteDetails `json:"site_details"`
//	}
func ScaffoldPagesDb(db *sql.DB) {
	// Create pages table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		lang TEXT NOT NULL DEFAULT 'en',
		slug TEXT NOT NULL UNIQUE,
		keywords TEXT NOT NULL DEFAULT '',
		author TEXT NOT NULL DEFAULT '',
		layout_name TEXT NOT NULL DEFAULT 'default',
		is_draft BOOLEAN NOT NULL DEFAULT 0,
		is_static BOOLEAN NOT NULL DEFAULT 0,
		require_auth BOOLEAN NOT NULL DEFAULT 0,
		required_roles TEXT NOT NULL DEFAULT '[]',
		file_path TEXT NOT NULL DEFAULT '',
		protected TEXT NOT NULL DEFAULT '',
		page_data TEXT NOT NULL DEFAULT '{}',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		published_at TIMESTAMP,
		meta_tags TEXT NOT NULL DEFAULT '[]'
	);
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Printf("Error creating pages table: %v", err)
		return
	}

	// Create index on slug for faster lookups
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_pages_slug ON pages(slug);")
	if err != nil {
		log.Printf("Error creating index on slug: %v", err)
	}

	// Create index on file_path for faster lookups
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_pages_file_path ON pages(file_path);")
	if err != nil {
		log.Printf("Error creating index on file_path: %v", err)
	}

	log.Println("Pages database tables scaffolded successfully")
}

// SQL queries for page operations

const (
	// InsertPageSQL inserts a new page into the database
	InsertPageSQL = `
	INSERT INTO pages (
		title, description, lang, slug, keywords, author,
		layout_name, is_draft, is_static, require_auth, required_roles,
		file_path, protected, page_data, created_at, updated_at, published_at, meta_tags
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	// UpdatePageSQL updates an existing page in the database by slug
	UpdatePageSQL = `
	UPDATE pages SET
		title = ?,
		description = ?,
		lang = ?,
		keywords = ?,
		author = ?,
		layout_name = ?,
		is_draft = ?,
		is_static = ?,
		require_auth = ?,
		required_roles = ?,
		file_path = ?,
		protected = ?,
		page_data = ?,
		updated_at = ?,
		published_at = ?,
		meta_tags = ?
	WHERE slug = ?;
	`

	// GetPageBySlugSQL retrieves a page by its slug
	GetPageBySlugSQL = `
	SELECT
		id, title, description, lang, slug, keywords, author,
		layout_name, is_draft, is_static, require_auth, required_roles,
		file_path, protected, page_data, created_at, updated_at, published_at, meta_tags
	FROM pages
	WHERE slug = ?;
	`

	// GetPageByFilePathSQL retrieves a page by its file path
	GetPageByFilePathSQL = `
	SELECT
		id, title, description, lang, slug, keywords, author,
		layout_name, is_draft, is_static, require_auth, required_roles,
		file_path, protected, page_data, created_at, updated_at, published_at, meta_tags
	FROM pages
	WHERE file_path = ?;
	`

	// ListPagesSQL lists all pages with pagination
	ListPagesSQL = `
	SELECT *
	FROM pages
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?;
	`

	// CountPagesSQL counts the total number of pages
	CountPagesSQL = `SELECT COUNT(*) FROM pages;`

	// DeletePageSQL deletes a page by its slug
	DeletePageSQL = `DELETE FROM pages WHERE slug = ?;`

	// SearchPagesSQL searches pages by title or content
	SearchPagesSQL = `
	SELECT
		id, title, description, lang, slug, keywords, author,
		layout_name, is_draft, is_static, require_auth, required_roles,
		file_path, protected, page_data, created_at, updated_at, published_at, meta_tags
	FROM pages
	WHERE title LIKE ? OR description LIKE ?
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?;
	`
)

// Helper functions for page operations

// InsertPage inserts a new page into the database
func InsertPage(db *sql.DB, page *models.Page) (int64, error) {
	// Handle potential nil slices
	if page.Keywords == nil {
		page.Keywords = []string{}
	}
	if page.RequiredRoles == nil {
		page.RequiredRoles = []string{}
	}

	keywords := strings.Join(page.Keywords, ",")
	requiredRoles := strings.Join(page.RequiredRoles, ",")

	// Convert page data to JSON string (in a real implementation, use proper JSON marshaling)
	pageData := "{}" // TODO: Convert page.PageData to JSON
	metaTags := "[]" // TODO: Convert page.MetaTags to JSON

	now := time.Now()
	page.CreatedAt = now
	page.UpdatedAt = now

	// Check if the page already exists by slug
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pages WHERE slug = ?)", page.Slug).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("error checking if page exists: %w", err)
	}

	// Insert new page
	result, err := db.Exec(
		InsertPageSQL,
		page.Title, page.Description, page.Lang, page.Slug, keywords, page.Author,
		page.LayoutName, page.IsDraft, page.IsStatic, page.RequireAuth, requiredRoles,
		page.FilePath, page.Protected, pageData, page.CreatedAt, page.UpdatedAt, page.PublishedAt, metaTags,
	)
	if err != nil {
		return 0, fmt.Errorf("error inserting page: %w", err)
	}

	return result.LastInsertId()
}

// UpdatePage updates an existing page in the database
func UpdatePage(db *sql.DB, page *models.Page) error {
	keywords := strings.Join(page.Keywords, ",")
	requiredRoles := strings.Join(page.RequiredRoles, ",")

	// Convert page data to JSON string (in a real implementation, use proper JSON marshaling)
	pageData := "{}" // TODO: Convert page.PageData to JSON
	metaTags := "[]" // TODO: Convert page.MetaTags to JSON

	page.UpdatedAt = time.Now()

	_, err := db.Exec(
		UpdatePageSQL,
		page.Title, page.Description, page.Lang, keywords, page.Author,
		page.LayoutName, page.IsDraft, page.IsStatic, page.RequireAuth, requiredRoles,
		page.FilePath, page.Protected, pageData, page.UpdatedAt, page.PublishedAt, metaTags,
		page.Slug,
	)
	if err != nil {
		return fmt.Errorf("error updating page: %w", err)
	}

	return nil
}

// GetPageBySlug retrieves a page by its slug
func GetPageBySlug(db *sql.DB, slug string) (*models.Page, error) {
	var page models.Page
	var keywords, requiredRoles, pageData, metaTags string
	var id int64
	var publishedAt sql.NullTime

	err := db.QueryRow(GetPageBySlugSQL, slug).Scan(
		&id, &page.Title, &page.Description, &page.Lang, &page.Slug, &keywords, &page.Author,
		&page.LayoutName, &page.IsDraft, &page.IsStatic, &page.RequireAuth, &requiredRoles,
		&page.FilePath, &page.Protected, &pageData, &page.CreatedAt, &page.UpdatedAt, &publishedAt, &metaTags,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting page by slug: %w", err)
	}

	// Handle nullable published_at date
	if publishedAt.Valid {
		page.PublishedAt = &publishedAt.Time
	}

	// Initialize slices to avoid nil
	page.Keywords = []string{}
	page.RequiredRoles = []string{}

	// Parse comma-separated strings into slices
	if keywords != "" {
		page.Keywords = strings.Split(keywords, ",")
	}

	if requiredRoles != "" {
		page.RequiredRoles = strings.Split(requiredRoles, ",")
	}

	// TODO: Parse pageData and metaTags from JSON
	page.PageData = make(map[string]string)

	return &page, nil
}

// DeletePage deletes a page by its slug
func DeletePage(db *sql.DB, slug string) error {
	_, err := db.Exec(DeletePageSQL, slug)
	if err != nil {
		return fmt.Errorf("error deleting page: %w", err)
	}

	return nil
}

// ListPages lists all pages with pagination
func ListPages(db *sql.DB, limit, offset int) ([]*models.Page, error) {
	rows, err := db.Query(ListPagesSQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing pages: %w", err)
	}
	defer rows.Close()

	var pages []*models.Page

	for rows.Next() {
		var page models.Page
		var keywords, requiredRoles, pageData, metaTags string
		var id int64
		var publishedAt sql.NullTime

		err := rows.Scan(
			&id, &page.Title, &page.Description, &page.Lang, &page.Slug, &keywords, &page.Author,
			&page.LayoutName, &page.IsDraft, &page.IsStatic, &page.RequireAuth, &requiredRoles,
			&page.FilePath, &page.Protected, &pageData, &page.CreatedAt, &page.UpdatedAt, &publishedAt, &metaTags,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning page row: %w", err)
		}

		// Handle nullable published_at date
		if publishedAt.Valid {
			page.PublishedAt = &publishedAt.Time
		}

		// Initialize slices to avoid nil
		page.Keywords = []string{}
		page.RequiredRoles = []string{}

		// Parse comma-separated strings into slices
		if keywords != "" {
			page.Keywords = strings.Split(keywords, ",")
		}

		if requiredRoles != "" {
			page.RequiredRoles = strings.Split(requiredRoles, ",")
		}

		// Initialize PageData map
		page.PageData = make(map[string]string)
		// TODO: Parse pageData and metaTags from JSON

		pages = append(pages, &page)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating page rows: %w", err)
	}

	return pages, nil
}

// CountPages counts the total number of pages
func CountPages(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(CountPagesSQL).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting pages: %w", err)
	}

	return count, nil
}

// SearchPages searches pages by title or description
func SearchPages(db *sql.DB, query string, limit, offset int) ([]*models.Page, error) {
	searchQuery := "%" + query + "%"
	rows, err := db.Query(SearchPagesSQL, searchQuery, searchQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error searching pages: %w", err)
	}
	defer rows.Close()

	var pages []*models.Page

	for rows.Next() {
		var page models.Page
		var keywords, requiredRoles, pageData, metaTags string
		var id int64
		var publishedAt sql.NullTime

		err := rows.Scan(
			&id, &page.Title, &page.Description, &page.Lang, &page.Slug, &keywords, &page.Author,
			&page.LayoutName, &page.IsDraft, &page.IsStatic, &page.RequireAuth, &requiredRoles,
			&page.FilePath, &page.Protected, &pageData, &page.CreatedAt, &page.UpdatedAt, &publishedAt, &metaTags,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning page row: %w", err)
		}

		// Handle nullable published_at date
		if publishedAt.Valid {
			page.PublishedAt = &publishedAt.Time
		}

		// Initialize slices to avoid nil
		page.Keywords = []string{}
		page.RequiredRoles = []string{}

		// Parse comma-separated strings into slices
		if keywords != "" {
			page.Keywords = strings.Split(keywords, ",")
		}

		if requiredRoles != "" {
			page.RequiredRoles = strings.Split(requiredRoles, ",")
		}

		// TODO: Parse pageData and metaTags from JSON
		page.PageData = make(map[string]string)

		pages = append(pages, &page)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating page rows: %w", err)
	}

	return pages, nil
}
