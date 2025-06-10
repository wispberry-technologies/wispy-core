package tests

import "fmt"

// Color codes for test output
const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGrey   = "\033[90m"
	colorReset  = "\033[0m"
)

func LogPass(testName string) string {
	return fmt.Sprintf("%s[Passed]%s %s%s%s", colorGreen, colorReset, colorGrey, testName, colorReset)
}

func LogFail(testName string) string {
	return fmt.Sprintf("%s[Failed]%s %s%s%s", colorRed, colorReset, colorGrey, testName, colorReset)
}

func LogWarn(msg string, args ...any) string {
	return fmt.Sprintf("%s[Warning]%s %s%s", colorYellow, colorReset, fmt.Sprintf(msg, args...), colorReset)
}
