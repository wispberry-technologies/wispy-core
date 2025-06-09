package common

import (
	"path/filepath"
)

// rootPath returns the root path for the Wispy core, appending any additional path segments
func RootPath(pathSegments ...string) string {
	return filepath.Join(append([]string{MustGetEnv("WISPY_CORE_ROOT")}, pathSegments...)...)
}

func RootSitesPath(pathSegments ...string) string {
	return filepath.Join(append([]string{MustGetEnv("WISPY_CORE_ROOT"), MustGetEnv("SITES_PATH")}, pathSegments...)...)
}
