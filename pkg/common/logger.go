package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// ColorHandler is a custom slog handler that adds colors and nice formatting
type ColorHandler struct {
	enableColor bool
	minLevel    slog.Level
}

func NewColorHandler(minLevel slog.Level) *ColorHandler {
	return &ColorHandler{
		enableColor: isTerminal(),
		minLevel:    minLevel,
	}
}

func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler
	return h
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf strings.Builder

	// Format timestamp
	if h.enableColor {
		buf.WriteString(Gray)
	}
	buf.WriteString(r.Time.Format("15:04:05"))
	if h.enableColor {
		buf.WriteString(Reset)
	}
	buf.WriteString(" ")

	// Format level with color
	var color string
	if h.enableColor {
		switch r.Level {
		case slog.LevelDebug:
			color = Gray
		case slog.LevelInfo:
			color = Blue
		case slog.LevelWarn:
			color = Yellow
		case slog.LevelError:
			color = Red
		default:
			color = White
		}
		buf.WriteString(color)
		buf.WriteString("[")
		buf.WriteString(r.Level.String())
		buf.WriteString("]")
		buf.WriteString(Reset)
	} else {
		buf.WriteString("[")
		buf.WriteString(r.Level.String())
		buf.WriteString("]")
	}
	buf.WriteString(" ")

	// Add message
	buf.WriteString(r.Message)

	// Add attributes in a clean format
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			buf.WriteString(" ")
			if h.enableColor {
				buf.WriteString(Gray)
			}
			buf.WriteString(a.Key)
			buf.WriteString("=")
			buf.WriteString(fmt.Sprintf("%v", a.Value.Any()))
			if h.enableColor {
				buf.WriteString(Reset)
			}
			return true
		})
	}

	buf.WriteString("\n")

	// Write to stdout
	_, err := os.Stdout.Write([]byte(buf.String()))
	return err
}

// isTerminal checks if we're running in a terminal that supports colors
func isTerminal() bool {
	// Check if we're in a terminal and not being piped
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Check for common CI environments that might not support colors
	if os.Getenv("CI") != "" || os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

// Global logger instance
var Log *slog.Logger

func init() {
	handler := NewColorHandler(slog.LevelDebug)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Convenience functions that support both printf-style and structured logging
func Debug(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		// Printf-style formatting
		Log.Debug(fmt.Sprintf(msg, args...))
	} else {
		// Structured logging
		Log.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		// Printf-style formatting
		Log.Info(fmt.Sprintf(msg, args...))
	} else {
		// Structured logging
		Log.Info(msg, args...)
	}
}

func Warning(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		// Printf-style formatting
		Log.Warn(fmt.Sprintf(msg, args...))
	} else {
		// Structured logging
		Log.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		// Printf-style formatting
		Log.Error(fmt.Sprintf(msg, args...))
	} else {
		// Structured logging
		Log.Error(msg, args...)
	}
}

// Helper function to detect if a string contains printf-style format verbs
func containsFormatVerbs(s string) bool {
	return strings.Contains(s, "%s") || strings.Contains(s, "%d") ||
		strings.Contains(s, "%v") || strings.Contains(s, "%f") ||
		strings.Contains(s, "%t") || strings.Contains(s, "%x") ||
		strings.Contains(s, "%X") || strings.Contains(s, "%o") ||
		strings.Contains(s, "%b") || strings.Contains(s, "%e") ||
		strings.Contains(s, "%E") || strings.Contains(s, "%g") ||
		strings.Contains(s, "%G") || strings.Contains(s, "%c") ||
		strings.Contains(s, "%q") || strings.Contains(s, "%U") ||
		strings.Contains(s, "%%")
}

// Special purpose logging functions - clean, no forced emojis
func Success(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		Log.Info(fmt.Sprintf(msg, args...), "log_type", "SUCCESS")
	} else {
		newArgs := append(args, "log_type", "SUCCESS")
		Log.Info(msg, newArgs...)
	}
}

func Startup(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		Log.Info(fmt.Sprintf(msg, args...), "log_type", "STARTUP")
	} else {
		newArgs := append(args, "log_type", "STARTUP")
		Log.Info(msg, newArgs...)
	}
}

func Progress(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		Log.Info(fmt.Sprintf(msg, args...), "log_type", "PROGRESS")
	} else {
		newArgs := append(args, "log_type", "PROGRESS")
		Log.Info(msg, newArgs...)
	}
}

func Config(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		Log.Info(fmt.Sprintf(msg, args...), "log_type", "CONFIG")
	} else {
		newArgs := append(args, "log_type", "CONFIG")
		Log.Info(msg, newArgs...)
	}
}

// Context-aware logging functions
func DebugCtx(ctx context.Context, msg string, args ...any) {
	Log.DebugContext(ctx, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	Log.InfoContext(ctx, msg, args...)
}

func WarningCtx(ctx context.Context, msg string, args ...any) {
	Log.WarnContext(ctx, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	Log.ErrorContext(ctx, msg, args...)
}

// Fatal logs an error and exits
func Fatal(msg string, args ...any) {
	if len(args) > 0 && containsFormatVerbs(msg) {
		Log.Error(fmt.Sprintf(msg, args...), "log_type", "FATAL")
	} else {
		newArgs := append(args, "log_type", "FATAL")
		Log.Error(msg, newArgs...)
	}
	os.Exit(1)
}

// Utility functions for direct color formatting
func Colorize(text string, color string) string {
	if handler, ok := Log.Handler().(*ColorHandler); ok && !handler.enableColor {
		return text
	}
	return color + text + Reset
}

func RedText(text string) string {
	return Colorize(text, Red)
}

func GreenText(text string) string {
	return Colorize(text, Green)
}

func YellowText(text string) string {
	return Colorize(text, Yellow)
}

func BlueText(text string) string {
	return Colorize(text, Blue)
}

func CyanText(text string) string {
	return Colorize(text, Cyan)
}

func GrayText(text string) string {
	return Colorize(text, Gray)
}

// Performance timing helpers using slog
type Timer struct {
	start time.Time
	name  string
}

func StartTimer(name string) *Timer {
	Debug("Starting timer: %s", name)
	return &Timer{
		start: time.Now(),
		name:  name,
	}
}

func (t *Timer) End() {
	duration := time.Since(t.start)
	Progress("Timer completed: %s in %v", t.name, duration)
}

func (t *Timer) EndWithMessage(msg string, args ...any) {
	duration := time.Since(t.start)
	// Add duration to the structured args
	newArgs := append(args, "duration", duration)
	Progress(msg, newArgs...)
}

// Log level management
func SetLogLevel(level slog.Level) {
	handler := NewColorHandler(level)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Disable colors (useful for CI/testing)
func DisableColors() {
	handler := NewColorHandler(slog.LevelDebug)
	handler.enableColor = false
	Log = slog.New(handler)
	slog.SetDefault(Log)
}
