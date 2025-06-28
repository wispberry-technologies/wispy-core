package common

import (
	"os"
)

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	// Get directory information
	info, err := os.Stat(path)
	if err == nil {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			return os.ErrExist
		}
		return nil
	}

	// Create directory with parent directories if needed
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	// Some other error occurred
	return err
}
