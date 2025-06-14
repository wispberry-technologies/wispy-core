package main

import (
	"fmt"
	"os"
	"wispy-core/core"
)

func finalTest() {
	// Create a template engine
	engine := core.NewTemplateEngine(core.DefaultFunctionMap())

	// Create context with test data
	ctx := core.NewTemplateContext(map[string]interface{}{
		"Page": map[string]interface{}{
			"Meta": map[string]interface{}{
				"Title":       "Home Page",
				"Description": "Welcome to our awesome site",
				"Author":      "Wispy Core Team",
				"CustomData": map[string]interface{}{
					"features_title":       "Amazing Features",
					"features_description": "Check out these great features",
				},
			},
		},
		"Site": map[string]interface{}{
			"Name": "Test Site",
		},
	}, engine)

	// First, process the home page to define blocks
	homePath := "/Users/theo/Desktop/wispy-core/sites/localhost/pages/home.html"
	if content, err := os.ReadFile(homePath); err == nil {
		_, homeErrors := engine.Render(string(content), ctx)
		if len(homeErrors) > 0 {
			fmt.Println("Home page errors:")
			for i, err := range homeErrors {
				fmt.Printf("%d. %v\n", i+1, err)
			}
		} else {
			fmt.Println("✅ Home page processed successfully!")
		}
	}

	// Now render the layout (which should use the defined blocks)
	layoutPath := "/Users/theo/Desktop/wispy-core/sites/localhost/layouts/default.html"
	if content, err := os.ReadFile(layoutPath); err == nil {
		result, errors := engine.Render(string(content), ctx)

		if len(errors) > 0 {
			fmt.Println("Layout errors:")
			for i, err := range errors {
				fmt.Printf("%d. %v\n", i+1, err)
			}
		} else {
			fmt.Println("✅ Layout rendered successfully!")
		}

		fmt.Printf("Result length: %d characters\n", len(result))

		// Check if the page-body block content is included
		if len(result) > 0 {
			// Look for content that should be from the home page
			if contains(result, "Lightning Fast") && contains(result, "Multisite Ready") {
				fmt.Println("✅ Page content correctly included in layout!")
			} else {
				fmt.Println("❌ Page content not found in layout")
			}

			// Check for meta tags
			if contains(result, "Welcome to our awesome site") {
				fmt.Println("✅ Meta description correctly rendered!")
			} else {
				fmt.Println("❌ Meta description not found")
			}
		}
	} else {
		fmt.Printf("Error reading layout: %v\n", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
