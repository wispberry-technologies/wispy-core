package common

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

// Global logger instance
var Log *slog.Logger

func init() {
	// Initialize with default text handler writing to stdout
	// handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 	Level: slog.LevelDebug,
	// })
	w := os.Stderr
	const LevelTrace = slog.LevelDebug - 4
	handler := tint.NewHandler(w, &tint.Options{
		NoColor: !isatty.IsTerminal(w.Fd()),
		Level:   LevelTrace,
	})
	Log = slog.New(handler)

	slog.SetDefault(Log)
}

// - Debug logs a debug message
func Debug(msg string, logArgs ...any) {
	fmt.Print(Reset, "[DEBUG] ", DarkGray)
	if strings.Contains(msg, "%s") {
		if len(logArgs) > 0 {
			fmt.Printf(msg, logArgs...)
		} else {
			fmt.Print(msg)
		}
		fmt.Println(Reset)
	} else {
		fmt.Print(msg)
		for i, arg := range logArgs {
			strArg, ok := arg.(string)
			if i%2 == 0 && i > 0 && ok {
				fmt.Print(DarkGray, strArg, "=", Reset)
			} else {
				fmt.Print(arg, " ")
			}
		}
		fmt.Println(Reset)
	}

}

// - Info logs an info message
func Info(msg string, logArgs ...any) {
	if strings.Contains(msg, "%") {
		if len(logArgs) > 0 {
			Log.Info(fmt.Sprintf(msg, logArgs...))
		}
	} else {
		Log.Info(msg, logArgs...)
	}
}

// - Warning logs a warning message
func Warning(msg string, args ...any) {
	if len(args) > 0 {
		Log.Warn(fmt.Sprintf(msg, args...))
	} else {
		Log.Warn(msg)
	}
}

// - Error logs an error message
func Error(msg string, args ...any) {
	if len(args) > 0 {
		Log.Error(fmt.Sprintf(msg, args...))
	} else {
		Log.Error(msg)
	}
}

// - Fatal logs an error message and exits
func Fatal(msg string, args ...any) {
	if len(args) > 0 {
		Log.Error(fmt.Sprintf(msg, args...))
	} else {
		Log.Error(msg)
	}
	os.Exit(1)
}

// - Success logs a success message using Info level
func Success(msg string, args ...any) {
	if len(args) > 0 {
		Log.Info(fmt.Sprintf(msg, args...), "type", "success")
	} else {
		Log.Info(msg, "type", "success")
	}
}

// - Startup logs a startup message using Info level
func ServerLog(msg string, args ...any) {
	if len(args) > 0 {
		Log.Info("type", "server", fmt.Sprintf(msg, args...))
	} else {
		Log.Info("type", "server", msg)
	}
}

// SetLogLevel changes the minimum log level
func SetLogLevel(level slog.Level) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	Log = slog.New(handler)
	slog.SetDefault(Log)
}
