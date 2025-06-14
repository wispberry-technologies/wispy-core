package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"wispy-core/tail"
)

func main() {
	// Get paths to example HTML and output CSS
	inputPath := filepath.Join("daisy-colors.html")
	outputPath := filepath.Join("daisy-colors-output.css")

	// Load HTML file
	htmlBytes, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read HTML file: %v", err)
	}

	// Compile the CSS
	css, err := tail.CompileHTML(string(htmlBytes))
	if err != nil {
		log.Fatalf("Failed to compile CSS: %v", err)
	}

	fmt.Printf("CSS generated successfully\n\n")

	// Write the CSS to file
	err = os.WriteFile(outputPath, []byte(css), 0644)
	if err != nil {
		log.Fatalf("Failed to write CSS file: %v", err)
	}

	fmt.Printf("CSS output written to %s\n", outputPath)
}
