package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"wispy-core/tail"
)

func main() {
	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Define input and output paths
	// Go up one level from current directory to find the example HTML file
	inputPath := filepath.Join(dir, "tail", "examples", "simple", "example-input.html")
	outputPath := filepath.Join(dir, "tail", "examples", "simple", "wispy-tail-output.css")

	// Read the HTML file
	html, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Compile HTML to CSS with daisyUI support
	css, err := tail.CompileHTMLWithDaisyUI(string(html))
	if err != nil {
		log.Fatalf("Error compiling HTML: %v", err)
	}

	// Write the CSS file
	err = os.WriteFile(outputPath, []byte(css), 0644)
	if err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	fmt.Printf("Successfully compiled Tailwind CSS to %s\n", outputPath)
}
