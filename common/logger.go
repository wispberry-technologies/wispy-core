package common

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Context   string    `json:"context,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	IP        string    `json:"ip,omitempty"`
	Method    string    `json:"method,omitempty"`
	URL       string    `json:"url,omitempty"`
	Status    int       `json:"status,omitempty"`
	Duration  int64     `json:"duration,omitempty"` // milliseconds
	File      string    `json:"file,omitempty"`
	Line      int       `json:"line,omitempty"`
}

// Logger provides centralized logging for the application
type Logger struct {
	siteDomain string
	dbPath     string
	db         *sql.DB
	minLevel   LogLevel
	slogger    *slog.Logger
}

// NewLogger creates a new logger for a specific site using SiteInstance
func NewLogger(instance *SiteInstance) (*Logger, error) {
	db, err := instance.GetDB("logs") // Ensure the logs database is initialized
	if err != nil {
		return nil, fmt.Errorf("failed to open logs database: %w", err)
	}

	// Create structured logger with site context
	slogger := slog.With("site", instance.Domain)

	logger := &Logger{
		siteDomain: instance.Domain,
		db:         db,
		minLevel:   DEBUG, // Default minimum level
		slogger:    slogger,
	}

	if err := logger.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize logs database: %w", err)
	}

	return logger, nil
}

// NewLoggerFromDomain creates a new logger for a specific site domain (fallback method)
// example: when creating a logger for a site that doesn't have an instance yet
func NewLoggerFromDomain(siteDomain, sitesBasePath string) (*Logger, error) {
	dbDir := filepath.Join(sitesBasePath, siteDomain, "dbs")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "logs.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open logs database: %w", err)
	}

	// Create structured logger with site context
	slogger := slog.With("site", siteDomain)

	logger := &Logger{
		siteDomain: siteDomain,
		dbPath:     dbPath,
		db:         db,
		minLevel:   INFO, // Default minimum level
		slogger:    slogger,
	}

	if err := logger.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize logs database: %w", err)
	}

	return logger, nil
}

// SetMinLevel sets the minimum log level to record
func (l *Logger) SetMinLevel(level LogLevel) {
	l.minLevel = level
}

// initDatabase creates the logs table if it doesn't exist
func (l *Logger) initDatabase() error {
	query := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		context TEXT,
		user_agent TEXT,
		ip TEXT,
		method TEXT,
		url TEXT,
		status INTEGER,
		duration INTEGER,
		file TEXT,
		line INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_status ON logs(status);
	`

	if _, err := l.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create logs table: %w", err)
	}

	return nil
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	// Get just the filename, not the full path
	return filepath.Base(file), line
}

// log writes a log entry to both console and database using slog
func (l *Logger) log(level LogLevel, message string, contextData map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	// Get caller information
	file, line := getCallerInfo(3) // Skip this method, the level method, and the public method

	// Create log entry for database
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		File:      file,
		Line:      line,
	}

	// Extract context information
	if contextData != nil {
		if userAgent, ok := contextData["user_agent"].(string); ok {
			entry.UserAgent = userAgent
		}
		if ip, ok := contextData["ip"].(string); ok {
			entry.IP = ip
		}
		if method, ok := contextData["method"].(string); ok {
			entry.Method = method
		}
		if url, ok := contextData["url"].(string); ok {
			entry.URL = url
		}
		if status, ok := contextData["status"].(int); ok {
			entry.Status = status
		}
		if duration, ok := contextData["duration"].(int64); ok {
			entry.Duration = duration
		}
		if ctx, ok := contextData["context"].(string); ok {
			entry.Context = ctx
		}
	}

	// Convert to slog level
	var slogLevel slog.Level
	switch level {
	case DEBUG:
		slogLevel = slog.LevelDebug
	case INFO:
		slogLevel = slog.LevelInfo
	case WARN:
		slogLevel = slog.LevelWarn
	case ERROR:
		slogLevel = slog.LevelError
	}

	// Build slog attributes
	attrs := []slog.Attr{
		slog.String("file", file),
		slog.Int("line", line),
	}

	if contextData != nil {
		for key, value := range contextData {
			switch v := value.(type) {
			case string:
				attrs = append(attrs, slog.String(key, v))
			case int:
				attrs = append(attrs, slog.Int(key, v))
			case int64:
				attrs = append(attrs, slog.Int64(key, v))
			case bool:
				attrs = append(attrs, slog.Bool(key, v))
			default:
				attrs = append(attrs, slog.String(key, fmt.Sprintf("%v", v)))
			}
		}
	}

	// Log using slog
	ctx := context.Background()
	l.slogger.LogAttrs(ctx, slogLevel, message, attrs...)

	// Also log to database
	l.logToDatabase(entry)
}

// logToDatabase stores the log entry in the database
func (l *Logger) logToDatabase(entry LogEntry) {
	if l.db == nil {
		// No database connection, skip database logging
		return
	}

	query := `
	INSERT INTO logs (timestamp, level, message, context, user_agent, ip, method, url, status, duration, file, line)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := l.db.Exec(query,
		entry.Timestamp,
		entry.Level.String(),
		entry.Message,
		entry.Context,
		entry.UserAgent,
		entry.IP,
		entry.Method,
		entry.URL,
		entry.Status,
		entry.Duration,
		entry.File,
		entry.Line,
	)

	if err != nil {
		// If database logging fails, use slog to log the error
		l.slogger.Error("Failed to write to database log", "error", err)
	}
}

// - Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(DEBUG, message, ctx)
}

// - Info logs an info message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(INFO, message, ctx)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(WARN, message, ctx)
}

// - Error logs an error message
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(ERROR, message, ctx)
}

// - LogHTTPRequest logs an HTTP request with timing information
func (l *Logger) LogHTTPRequest(method, url, userAgent, ip string, status int, duration time.Duration) {
	context := map[string]interface{}{
		"method":     method,
		"url":        url,
		"user_agent": userAgent,
		"ip":         ip,
		"status":     status,
		"duration":   duration.Milliseconds(),
		"context":    "HTTP",
	}

	level := INFO
	if status >= 400 && status < 500 {
		level = WARN
	} else if status >= 500 {
		level = ERROR
	}

	message := fmt.Sprintf("HTTP Request: %s %s", method, url)
	l.log(level, message, context)
}

// GetLogs retrieves log entries from the database with optional filtering
func (l *Logger) GetLogs(limit int, level LogLevel, since *time.Time) ([]LogEntry, error) {
	var query strings.Builder
	var args []interface{}

	query.WriteString(`
		SELECT id, timestamp, level, message, context, user_agent, ip, method, url, status, duration, file, line
		FROM logs WHERE 1=1
	`)

	if level != DEBUG {
		query.WriteString(" AND level IN (")
		levelNames := []string{}
		for l := level; l <= ERROR; l++ {
			levelNames = append(levelNames, "'"+l.String()+"'")
		}
		query.WriteString(strings.Join(levelNames, ","))
		query.WriteString(")")
	}

	if since != nil {
		query.WriteString(" AND timestamp >= ?")
		args = append(args, since)
	}

	query.WriteString(" ORDER BY timestamp DESC")

	if limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, limit)
	}

	rows, err := l.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var levelStr string
		var userAgent, ip, method, url, context, file sql.NullString
		var status, line sql.NullInt64

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&levelStr,
			&entry.Message,
			&context,
			&userAgent,
			&ip,
			&method,
			&url,
			&status,
			&entry.Duration,
			&file,
			&line,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Parse level
		switch levelStr {
		case "DEBUG":
			entry.Level = DEBUG
		case "INFO":
			entry.Level = INFO
		case "WARN":
			entry.Level = WARN
		case "ERROR":
			entry.Level = ERROR
		}

		// Handle nullable fields
		if userAgent.Valid {
			entry.UserAgent = userAgent.String
		}
		if ip.Valid {
			entry.IP = ip.String
		}
		if method.Valid {
			entry.Method = method.String
		}
		if url.Valid {
			entry.URL = url.String
		}
		if context.Valid {
			entry.Context = context.String
		}
		if file.Valid {
			entry.File = file.String
		}
		if status.Valid {
			entry.Status = int(status.Int64)
		}
		if line.Valid {
			entry.Line = int(line.Int64)
		}

		logs = append(logs, entry)
	}

	return logs, nil
}
