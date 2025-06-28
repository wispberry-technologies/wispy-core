package main

import (
	"fmt"
	"log"
	"path/filepath"

	"wispy-core/core/site"
)

// contains checks if a string is in a slice of strings
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func main() {

	// Create the scaffold configuration
	scaffoldConfig := site.ScaffoldConfig{
		ID:               "3123123123",
		Name:             "test",
		Domain:           "abc.co",
		BaseURL:          "http://abc.co",
		ThemeName:        "default",
		ThemeMode:        "light", // Default to light, can be overridden with a flag
		ContentTypes:     []string{"page", "post"},
		WithExample:      true,
		ThemeOptions:     make(map[string]string),
		TypographyPreset: "default",
		ColorPreset:      "default",
	}

	absOutputDir, err := filepath.Abs("./_data/tenants/")
	if err != nil {
		log.Fatalf("Error resolving output directory path: %v", err)
	}

	// Generate the site scaffold
	fmt.Printf("Generating site '%s' in directory: %s\n", scaffoldConfig.Name, absOutputDir)
	_, err = site.Scaffold(absOutputDir, scaffoldConfig)
	if err != nil {
		log.Fatalf("Error generating site: %v", err)
	}

	fmt.Println("✅ Site generated successfully!")
	fmt.Println("")
	fmt.Println("📂 Site directory: " + absOutputDir)
	fmt.Println("🌐 Site URL: http://" + scaffoldConfig.Domain + ".localhost:3000")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", absOutputDir)
	fmt.Println("  2. wispy serve")
	fmt.Println("")
	fmt.Println("You can customize your theme in: " + filepath.Join(absOutputDir, "assets", "css"))
	fmt.Println("Add content in: " + filepath.Join(absOutputDir, "content"))
}
