package common

import (
	"os"
	"strconv"
)

// getEnv gets an environment variable with a fallback default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnv gets an environment variable with a fallback default
func MustGetEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic("Environment variable " + key + " is not set!")
}

// getEnvInt gets an environment variable as an integer with a fallback default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as a boolean with a fallback default
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func IsProduction() bool {
	isProd := GetEnv("ENN", "development")
	if isProd == "PRODUCTION" || isProd == "PROD" || isProd == "production" || isProd == "prod" {
		return true
	}
	return false
}
