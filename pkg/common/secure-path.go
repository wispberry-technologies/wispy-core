package common

import (
	"path/filepath"
)

// rootPath returns the root path for the Wispy core, appending any additional path segments
func RootPath(pathSegments ...string) string {
	return filepath.Join(append([]string{MustGetEnv("WISPY_CORE_ROOT")}, pathSegments...)...)
}

// sitesBasePath caches the computed sites base path
var sitesBasePath string

func RootSitesPath(pathSegments ...string) string {
	// Initialize sitesBasePath if it's empty
	if sitesBasePath == "" {
		coreRoot := MustGetEnv("WISPY_CORE_ROOT")
		sitesPath := MustGetEnv("SITES_PATH")
		
		// Handle relative or absolute paths correctly
		if filepath.IsAbs(sitesPath) {
			// If SITES_PATH is absolute, use it directly
			sitesBasePath = sitesPath
		} else {
			// If SITES_PATH is relative, join with WISPY_CORE_ROOT
			sitesBasePath = filepath.Join(coreRoot, sitesPath)
		}
		Info("Sites base path set to: %s", sitesBasePath)
	}

	// Add any additional path segments
	return filepath.Join(append([]string{sitesBasePath}, pathSegments...)...)
}
