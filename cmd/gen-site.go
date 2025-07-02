package main

import (
	"fmt"
	"log"
	"path/filepath"

	"wispy-core/core/site"
)

func main() {

	// Create the scaffold configuration
	scaffoldConfig := site.ScaffoldConfig{
		Name:         "test",
		Domain:       "abc.co",
		BaseURL:      "http://abc.co",
		ContentTypes: []string{"page", "post"},
		WithExample:  true,
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
