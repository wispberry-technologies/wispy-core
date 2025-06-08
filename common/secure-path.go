package common

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	// get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}
	// Set the WISPY_CORE_ROOT environment variable to the current directory
	os.Setenv("WISPY_CORE_ROOT", currentDir)

	// Ensure the root path exists
	if err := os.MkdirAll(rootPath(), 0755); err != nil {
		panic("Failed to create root path: " + err.Error())
	}
}

// rootPath returns the root path for the Wispy core, appending any additional path segments
func rootPath(pathSegments ...string) string {
	return filepath.Join(append([]string{os.Getenv("WISPY_CORE_ROOT")}, pathSegments...)...)
}

// SitesPath returns the secure path to the sites directory
func SitesPath(pathSegments ...string) string {
	sitesPath := MustGetEnv("SITES_PATH")
	return filepath.Join(append([]string{sitesPath}, pathSegments...)...)
}

func ValidateSitePath(siteDomain string, pathSegments ...string) (string, error) {
	relativePath := filepath.Join(MustGetEnv("SITES_PATH"), siteDomain, filepath.Join(pathSegments...))
	return ValidatePath(relativePath)
}

// SecureReadFile safely reads a file within the allowed base path
func SecureReadFile(relativePath string) ([]byte, error) {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(safePath)
}

// SecureWriteFile safely writes a file within the allowed base path
func SecureWriteFile(relativePath string, data []byte, perm os.FileMode) error {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return err
	}

	return os.WriteFile(safePath, data, perm)
}

// SecureMkdirAll safely creates directories within the allowed base path
func SecureMkdirAll(relativePath string, perm os.FileMode) error {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return err
	}

	return os.MkdirAll(safePath, perm)
}

// SecureStat safely gets file info within the allowed base path
func SecureStat(relativePath string) (os.FileInfo, error) {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return nil, err
	}

	return os.Stat(safePath)
}

// SecureExists checks if a file exists within the allowed base path
func SecureExists(relativePath string) bool {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return false
	}

	_, err = os.Stat(safePath)
	return err == nil
}

// SecureWalk safely walks a directory tree within the allowed base path
func SecureWalk(relativePath string, walkFn filepath.WalkFunc) error {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return err
	}

	return filepath.Walk(safePath, walkFn)
}

// SecureServeFile safely serves a file within the allowed base path (returns the safe path for http.ServeFile)
func SecureServeFile(relativePath string) (string, error) {
	return ValidatePath(relativePath)
}

// SecureGlob safely returns file paths matching a pattern within the allowed base path
func SecureGlob(pattern string) ([]string, error) {
	// First validate the pattern path
	_, err := ValidatePath(pattern)
	if err != nil {
		return nil, err
	}

	// Build the full pattern path
	fullPattern := filepath.Join(rootPath(), pattern)

	// Get all matches
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}

	// Validate each match to ensure it's still within our base path
	var validMatches []string
	for _, match := range matches {
		// Convert back to relative path for validation
		relPath, err := filepath.Rel(rootPath(), match)
		if err != nil {
			continue // Skip this match if we can't get relative path
		}

		// Validate the relative path
		if _, err := ValidatePath(relPath); err != nil {
			continue // Skip this match if it fails validation
		}

		validMatches = append(validMatches, match)
	}

	return validMatches, nil
}

// SecureRemove safely removes a file within the allowed base path
func SecureRemove(relativePath string) error {
	safePath, err := ValidatePath(relativePath)
	if err != nil {
		return err
	}

	return os.Remove(safePath)
}

// validatePath validates and sanitizes a path to prevent traversal attacks
func ValidatePath(relativePath string) (string, error) {
	// Clean the path to remove any redundant separators or elements
	cleanPath := filepath.Clean(relativePath)

	// Reject absolute paths
	if filepath.IsAbs(cleanPath) {
		return "", errors.New("absolute paths are not allowed")
	}

	// Reject paths that contain ".." or start with ".."
	if strings.Contains(cleanPath, "..") {
		return "", errors.New("path traversal detected")
	}

	// Join with base path
	fullPath := filepath.Join(rootPath(), cleanPath)

	// Ensure the resulting path is still within the base path
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	absBasePath, err := filepath.Abs(rootPath())
	if err != nil {
		return "", err
	}

	// Check if the absolute full path starts with the absolute base path
	if !strings.HasPrefix(absFullPath, absBasePath+string(filepath.Separator)) && absFullPath != absBasePath {
		return "", errors.New("path outside allowed directory")
	}

	return fullPath, nil
}
