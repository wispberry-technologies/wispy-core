package app

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"wispy-core/common"
	"wispy-core/config"

	_ "github.com/mattn/go-sqlite3"
)

// CMSDataProvider provides data for CMS dashboard and pages
type CMSDataProvider struct {
	domain string
}

// NewCMSDataProvider creates a new CMS data provider
func NewCMSDataProvider(domain string) *CMSDataProvider {
	return &CMSDataProvider{
		domain: domain,
	}
}

// DashboardStats represents statistics for the dashboard
type DashboardStats struct {
	FormCount       int `json:"formCount"`
	SubmissionCount int `json:"submissionCount"`
	SettingsCount   int `json:"settingsCount"`
}

// FormsStats represents statistics for the forms page
type FormsStats struct {
	TotalForms       int `json:"totalForms"`
	SubmissionsToday int `json:"submissionsToday"`
	ResponseRate     int `json:"responseRate"`
}

// SubmissionsStats represents statistics for the submissions page
type SubmissionsStats struct {
	TotalSubmissions  int    `json:"totalSubmissions"`
	WeeklySubmissions int    `json:"weeklySubmissions"`
	WeeklyChange      string `json:"weeklyChange"`
	UnreadSubmissions int    `json:"unreadSubmissions"`
	SpamFiltered      int    `json:"spamFiltered"`
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// FormItem represents a form in the forms list
type FormItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Submissions int       `json:"submissions"`
	Status      string    `json:"status"`
	Created     string    `json:"created"`
	CreatedAt   time.Time `json:"createdAt"`
}

// SubmissionItem represents a submission in the submissions list
type SubmissionItem struct {
	ID          string `json:"id"`
	FormName    string `json:"form"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
	Date        string `json:"date"`
	TimeAgo     string `json:"timeAgo"`
	Status      string `json:"status"`
	StatusStyle string `json:"statusStyle"`
	Initials    string `json:"initials"`
}

// GetFormsDB returns the forms database connection
func (dp *CMSDataProvider) GetFormsDB() (*sql.DB, error) {
	gConfig := config.GetGlobalConfig()
	dbPath := filepath.Join(gConfig.GetProjectRoot(), "_data", "tenants", dp.domain, "databases", "forms.db")

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := common.EnsureDir(dbDir); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open forms database: %w", err)
	}

	// Initialize database if it doesn't exist
	if err := dp.initializeFormsDB(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize forms database: %w", err)
	}

	return db, nil
}

// initializeFormsDB initializes the forms database with tables if they don't exist
func (dp *CMSDataProvider) initializeFormsDB(db *sql.DB) error {
	// Create forms table
	formsTableSQL := `
    CREATE TABLE IF NOT EXISTS forms (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uuid TEXT NOT NULL UNIQUE,
        name TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        description TEXT,
        fields TEXT NOT NULL,
        settings TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	// Create form submissions table
	submissionsTableSQL := `
    CREATE TABLE IF NOT EXISTS form_submissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uuid TEXT NOT NULL UNIQUE,
        form_id INTEGER NOT NULL,
        first_name TEXT,
        last_name TEXT,
        tags TEXT,
        email TEXT NOT NULL,
        tel TEXT,
        subject TEXT,
        message TEXT,
        data TEXT NOT NULL,
        ip_address TEXT,
        user_agent TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
    );`

	// Create indexes
	indexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_forms_uuid ON forms(uuid);`,
		`CREATE INDEX IF NOT EXISTS idx_forms_name ON forms(name);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_form_id ON form_submissions(form_id);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_email ON form_submissions(email);`,
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

// GetDashboardStats returns statistics for the dashboard
func (dp *CMSDataProvider) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	db, err := dp.GetFormsDB()
	if err != nil {
		return nil, err
	}

	stats := &DashboardStats{}

	// Get form count
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forms").Scan(&stats.FormCount)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get form count: %v", err)
		stats.FormCount = 0
	}

	// Get submission count
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM form_submissions").Scan(&stats.SubmissionCount)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get submission count: %v", err)
		stats.SubmissionCount = 0
	}

	// Settings count is relatively static for now
	stats.SettingsCount = 8

	return stats, nil
}

// GetFormsStats returns statistics for the forms page
func (dp *CMSDataProvider) GetFormsStats(ctx context.Context) (*FormsStats, error) {
	db, err := dp.GetFormsDB()
	if err != nil {
		return nil, err
	}

	stats := &FormsStats{}

	// Get total forms count
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forms").Scan(&stats.TotalForms)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get forms count: %v", err)
		stats.TotalForms = 0
	}

	// Get submissions today
	today := time.Now().Format("2006-01-02")
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM form_submissions WHERE DATE(created_at) = ?",
		today,
	).Scan(&stats.SubmissionsToday)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get today's submissions: %v", err)
		stats.SubmissionsToday = 0
	}

	// Calculate response rate (simplified - percentage of forms with submissions)
	if stats.TotalForms > 0 {
		var formsWithSubmissions int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT f.id) 
			FROM forms f 
			INNER JOIN form_submissions fs ON f.id = fs.form_id
		`).Scan(&formsWithSubmissions)
		if err == nil {
			stats.ResponseRate = (formsWithSubmissions * 100) / stats.TotalForms
		} else {
			stats.ResponseRate = 0
		}
	} else {
		stats.ResponseRate = 0
	}

	return stats, nil
}

// GetSubmissionsStats returns statistics for the submissions page
func (dp *CMSDataProvider) GetSubmissionsStats(ctx context.Context) (*SubmissionsStats, error) {
	db, err := dp.GetFormsDB()
	if err != nil {
		return nil, err
	}

	stats := &SubmissionsStats{}

	// Get total submissions
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM form_submissions").Scan(&stats.TotalSubmissions)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get total submissions: %v", err)
		stats.TotalSubmissions = 0
	}

	// Get this week's submissions
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM form_submissions WHERE DATE(created_at) >= ?",
		weekStart,
	).Scan(&stats.WeeklySubmissions)
	if err != nil && err != sql.ErrNoRows {
		common.Error("Failed to get weekly submissions: %v", err)
		stats.WeeklySubmissions = 0
	}

	// Get last week's submissions for comparison
	lastWeekStart := time.Now().AddDate(0, 0, -7-int(time.Now().Weekday())).Format("2006-01-02")
	lastWeekEnd := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
	var lastWeekSubmissions int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM form_submissions WHERE DATE(created_at) >= ? AND DATE(created_at) < ?",
		lastWeekStart, lastWeekEnd,
	).Scan(&lastWeekSubmissions)
	if err == nil && lastWeekSubmissions > 0 {
		change := ((stats.WeeklySubmissions - lastWeekSubmissions) * 100) / lastWeekSubmissions
		if change > 0 {
			stats.WeeklyChange = fmt.Sprintf("↗︎ %d%%", change)
		} else if change < 0 {
			stats.WeeklyChange = fmt.Sprintf("↘︎ %d%%", -change)
		} else {
			stats.WeeklyChange = "→ 0%"
		}
	} else {
		stats.WeeklyChange = "↗︎ New"
	}

	// For now, all submissions are considered "read" by default
	// In a real implementation, you'd add a "read" column to track this
	stats.UnreadSubmissions = 0

	// Spam filtering would require additional logic/integration
	stats.SpamFiltered = 0

	return stats, nil
}

// GetRecentActivity returns recent activity items for the dashboard
func (dp *CMSDataProvider) GetRecentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	activities := []ActivityItem{}

	// Get recent form submissions
	db, err := dp.GetFormsDB()
	if err != nil {
		return activities, err
	}

	rows, err := db.QueryContext(ctx, `
		SELECT fs.email, fs.created_at, f.title 
		FROM form_submissions fs 
		JOIN forms f ON fs.form_id = f.id 
		ORDER BY fs.created_at DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		common.Error("Failed to get recent submissions: %v", err)
		return activities, nil
	}
	defer rows.Close()

	for rows.Next() {
		var email, formTitle string
		var createdAt time.Time
		if err := rows.Scan(&email, &createdAt, &formTitle); err != nil {
			continue
		}

		activities = append(activities, ActivityItem{
			Title:       fmt.Sprintf("New submission from %s", email),
			Description: fmt.Sprintf("Submitted %s form", formTitle),
			Timestamp:   formatTimeAgo(createdAt),
			Icon:        "SUB",
			Color:       "primary",
		})
	}

	// If no activities, add a default system message
	if len(activities) == 0 {
		activities = append(activities, ActivityItem{
			Title:       "System ready",
			Description: "CMS is ready for use",
			Timestamp:   "Just now",
			Icon:        "SYS",
			Color:       "neutral",
		})
	}

	return activities, nil
}

// GetForms returns a list of forms
func (dp *CMSDataProvider) GetForms(ctx context.Context, limit int) ([]FormItem, error) {
	db, err := dp.GetFormsDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT f.uuid, f.name, f.title, f.description, f.created_at,
		       COUNT(fs.id) as submission_count
		FROM forms f
		LEFT JOIN form_submissions fs ON f.id = fs.form_id
		GROUP BY f.id, f.uuid, f.name, f.title, f.description, f.created_at
		ORDER BY f.created_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	var forms []FormItem
	for rows.Next() {
		var form FormItem
		var description sql.NullString

		err := rows.Scan(
			&form.ID,
			&form.Name,
			&form.Title,
			&description,
			&form.CreatedAt,
			&form.Submissions,
		)
		if err != nil {
			common.Error("Failed to scan form: %v", err)
			continue
		}

		if description.Valid {
			form.Description = description.String
		}

		form.Status = "Active" // Simplified for now
		form.Created = form.CreatedAt.Format("2006-01-02")

		forms = append(forms, form)
	}

	return forms, nil
}

// GetSubmissions returns a list of submissions
func (dp *CMSDataProvider) GetSubmissions(ctx context.Context, formFilter, statusFilter string, limit int) ([]SubmissionItem, error) {
	db, err := dp.GetFormsDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT fs.uuid, f.name, fs.first_name, fs.last_name, fs.email, 
		       fs.subject, fs.message, fs.created_at
		FROM form_submissions fs
		JOIN forms f ON fs.form_id = f.id
		WHERE 1=1
	`

	args := []interface{}{}

	// Add filters
	if formFilter != "" {
		query += " AND f.name = ?"
		args = append(args, formFilter)
	}

	query += " ORDER BY fs.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()

	var submissions []SubmissionItem
	for rows.Next() {
		var submission SubmissionItem
		var firstName, lastName, subject, message sql.NullString
		var createdAt time.Time

		err := rows.Scan(
			&submission.ID,
			&submission.FormName,
			&firstName,
			&lastName,
			&submission.Email,
			&subject,
			&message,
			&createdAt,
		)
		if err != nil {
			common.Error("Failed to scan submission: %v", err)
			continue
		}

		// Build name from first and last name
		name := ""
		if firstName.Valid && lastName.Valid {
			name = fmt.Sprintf("%s %s", firstName.String, lastName.String)
		} else if firstName.Valid {
			name = firstName.String
		} else if lastName.Valid {
			name = lastName.String
		} else {
			name = "Anonymous"
		}
		submission.Name = name

		// Generate initials
		if firstName.Valid && lastName.Valid {
			submission.Initials = fmt.Sprintf("%s%s",
				strings.ToUpper(string(firstName.String[0])),
				strings.ToUpper(string(lastName.String[0])))
		} else if firstName.Valid && len(firstName.String) > 0 {
			submission.Initials = strings.ToUpper(string(firstName.String[0]))
		} else {
			submission.Initials = "A"
		}

		if subject.Valid {
			submission.Subject = subject.String
		}
		if message.Valid {
			submission.Message = message.String
		}

		submission.Date = createdAt.Format("Jan 02, 2006")
		submission.TimeAgo = formatTimeAgo(createdAt)
		submission.Status = "New" // Simplified for now
		submission.StatusStyle = "badge-warning"

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// formatTimeAgo formats a time as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}
