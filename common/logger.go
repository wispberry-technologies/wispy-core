package common

import (
	"database/sql"
	"fmt"
	"log"
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
}

// NewLogger creates a new logger for a specific site
func NewLogger(siteDomain string, sitesBasePath string) (*Logger, error) {
	dbDir := filepath.Join(sitesBasePath, siteDomain, "dbs")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "logs.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open logs database: %w", err)
	}

	logger := &Logger{
		siteDomain: siteDomain,
		dbPath:     dbPath,
		db:         db,
		minLevel:   INFO, // Default minimum level
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

// log writes a log entry to both console and database
func (l *Logger) log(level LogLevel, message string, context map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	// Get caller information
	file, line := getCallerInfo(3) // Skip this method, the level method, and the public method

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		File:      file,
		Line:      line,
	}

	// Extract context information
	if context != nil {
		if userAgent, ok := context["user_agent"].(string); ok {
			entry.UserAgent = userAgent
		}
		if ip, ok := context["ip"].(string); ok {
			entry.IP = ip
		}
		if method, ok := context["method"].(string); ok {
			entry.Method = method
		}
		if url, ok := context["url"].(string); ok {
			entry.URL = url
		}
		if status, ok := context["status"].(int); ok {
			entry.Status = status
		}
		if duration, ok := context["duration"].(int64); ok {
			entry.Duration = duration
		}
		if ctx, ok := context["context"].(string); ok {
			entry.Context = ctx
		}
	}

	// Log to console with colors
	l.logToConsole(entry)

	// Log to database
	l.logToDatabase(entry)
}

// logToConsole outputs the log entry to console with colors
func (l *Logger) logToConsole(entry LogEntry) {
	var levelColor string
	switch entry.Level {
	case DEBUG:
		levelColor = "\033[36m" // Cyan
	case INFO:
		levelColor = "\033[32m" // Green
	case WARN:
		levelColor = "\033[33m" // Yellow
	case ERROR:
		levelColor = "\033[31m" // Red
	}

	reset := "\033[0m"
	gray := "\033[90m"

	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Format: [TIMESTAMP] [SITE] LEVEL: MESSAGE [context] (file:line)
	var contextStr string
	if entry.Context != "" {
		contextStr = fmt.Sprintf(" [%s]", entry.Context)
	}

	locationStr := ""
	if entry.File != "" {
		locationStr = fmt.Sprintf(" (%s:%d)", entry.File, entry.Line)
	}

	logLine := fmt.Sprintf("%s[%s] [%s] %s%s%s: %s%s%s%s%s",
		gray, timestamp, l.siteDomain, levelColor, entry.Level.String(), reset,
		entry.Message, contextStr, gray, locationStr, reset)

	// For HTTP requests, add method, URL, status, and duration
	if entry.Method != "" && entry.URL != "" {
		httpInfo := fmt.Sprintf(" %s%s %s %d (%dms)%s",
			gray, entry.Method, entry.URL, entry.Status, entry.Duration, reset)
		logLine += httpInfo
	}

	log.Println(logLine)
}

// logToDatabase stores the log entry in the database
func (l *Logger) logToDatabase(entry LogEntry) {
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
		// If database logging fails, at least log to console
		log.Printf("Failed to write to database log: %v", err)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(DEBUG, message, ctx)
}

// Info logs an info message
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

// Error logs an error message
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(ERROR, message, ctx)
}

// LogHTTPRequest logs an HTTP request with timing information
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

// Close closes the database connection
func (l *Logger) Close() error {
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

// LoggerManager manages loggers for multiple sites
type LoggerManager struct {
	loggers       map[string]*Logger
	sitesBasePath string
}

// NewLoggerManager creates a new logger manager
func NewLoggerManager(sitesBasePath string) *LoggerManager {
	return &LoggerManager{
		loggers:       make(map[string]*Logger),
		sitesBasePath: sitesBasePath,
	}
}

// GetLogger gets or creates a logger for a site
func (lm *LoggerManager) GetLogger(siteDomain string) (*Logger, error) {
	if logger, exists := lm.loggers[siteDomain]; exists {
		return logger, nil
	}

	logger, err := NewLogger(siteDomain, lm.sitesBasePath)
	if err != nil {
		return nil, err
	}

	lm.loggers[siteDomain] = logger
	return logger, nil
}

// GetDefaultLogger returns a console-only logger for system-level logging
func (lm *LoggerManager) GetDefaultLogger() *Logger {
	return &Logger{
		siteDomain: "system",
		db:         nil, // No database logging for default logger
	}
}

// CloseAll closes all loggers
func (lm *LoggerManager) CloseAll() error {
	var lastErr error
	for _, logger := range lm.loggers {
		if err := logger.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
