package common

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
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

func GetEnvOrSet(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// Set the environment variable if it doesn't exist
	if err := os.Setenv(key, defaultValue); err != nil {
		panic("Failed to set environment variable " + key + ": " + err.Error())
	}
	return defaultValue
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
	isProd := GetEnv("ENV", "local")
	if isProd == "PRODUCTION" || isProd == "PROD" || isProd == "production" || isProd == "prod" {
		return true
	}
	return false
}

func IsLocalDevelopment() bool {
	env := GetEnv("ENV", "local")
	return env == "LOCAL" || env == "local"
}

func IsStaging() bool {
	env := GetEnv("ENV", "local")
	return env == "staging" || env == "STAGING"
}

func LoadDotEnv() {
	// Attempt to load .env file from current directory or project root
	err := godotenv.Load(".env")

	if err != nil {
		// Try looking in parent directory (if running from /server)
		err = godotenv.Load("../.env")
		if err != nil {
			Fatal("Error loading .env file: %v", err)
		}
	}
}
