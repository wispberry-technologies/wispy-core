package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Base paths
	goPackages   = "/opt/homebrew/Cellar/go/1.24.2/libexec/src"
	destTemplate = "./pkg/"
)

// Map of source templates to destination folder names
var templatePaths = map[string]string{
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/text/template":     "go_templates/textTemplate",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/html/template":     "go_templates/htmlTemplate",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/fmtsort":  "go_templates/internal/fmtsort",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/testenv":  "go_templates/internal/testenv",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/cfg":      "go_templates/internal/cfg",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/goarch":   "go_templates/internal/goarch",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/platform": "go_templates/internal/platform",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/diff":     "go_templates/internal/diff",
	"/opt/homebrew/Cellar/go/1.24.2/libexec/src/internal/txtar":    "go_templates/internal/txtar",
}

// Import replacements
type importReplacement struct {
	from string
	to   string
}

var replacements = []importReplacement{
	// Imports
	{`"text/template/parse"`, `"wispy-core/pkg/go_templates/textTemplate/parse"`},
	{`"text/template"`, `"wispy-core/pkg/go_templates/textTemplate"`},
	{`"html/template"`, `"wispy-core/pkg/go_templates/htmlTemplate"`},
	{`"internal/fmtsort"`, `"wispy-core/pkg/go_templates/internal/fmtsort"`},
	{`"internal/testenv"`, `"wispy-core/pkg/go_templates/internal/testenv"`},
	{`"internal/cfg"`, `"wispy-core/pkg/go_templates/internal/cfg"`},
	{`"internal/goarch"`, `"wispy-core/pkg/go_templates/internal/goarch"`},
	{`"internal/platform"`, `"wispy-core/pkg/go_templates/internal/platform"`},
	{`"internal/diff"`, `"wispy-core/pkg/go_templates/internal/diff"`},
	{`"internal/txtar"`, `"wispy-core/pkg/go_templates/internal/txtar"`},
	// Other replacements
	{`	"internal/godebug"`, ` //	"internal/godebug"`},
	{`var debugAllowActionJSTmpl = godebug.New("jstmpllitinterp")`, `// var debugAllowActionJSTmpl = godebug.New("jstmpllitinterp")`},
}

func main() {
	// Create destination directory if it doesn't exist
	err := os.MkdirAll(destTemplate, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		os.Exit(1)
	}

	// Process each template
	destPaths := make([]string, 0, len(templatePaths))

	for srcPath, destFolder := range templatePaths {
		// Check if source directory exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("Error: Source directory '%s' not found.\n", srcPath)
			os.Exit(1)
		}

		// Calculate destination path
		destPath := filepath.Join(destTemplate, destFolder)
		destPaths = append(destPaths, destPath)

		// Remove destination directory if it exists
		os.RemoveAll(destPath)

		// Copy template package
		fmt.Printf("Copying %s to %s\n", srcPath, destPath)
		err = copyDirectory(srcPath, destPath)
		if err != nil {
			fmt.Printf("Error copying template: %v\n", err)
			os.Exit(1)
		}

		// Make directory writable
		err = os.Chmod(destPath, 0755)
		if err != nil {
			fmt.Printf("Warning: Could not change directory permissions: %v\n", err)
		}
	}

	fmt.Println("Successfully copied all template packages to their destinations")

	// Replace imports in all template packages
	for _, destPath := range destPaths {
		err = replaceImportsInDirectory(destPath)
		if err != nil {
			fmt.Printf("Error replacing imports in %s: %v\n", destPath, err)
			os.Exit(1)
		}
	}

	fmt.Println("Successfully replaced imports in template packages")
	fmt.Println("Import replacements:")
	for _, rep := range replacements {
		if !strings.HasPrefix(rep.from, "//") {
			fmt.Printf("  - %s → %s\n", rep.from, rep.to)
		}
	}
}

// copyDirectory copies a directory and its contents recursively
func copyDirectory(src, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Read the directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursive copy for subdirectories
			err = copyDirectory(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy files
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}

			// Make sure the file permissions allow writing
			err = os.WriteFile(dstPath, data, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// replaceImportsInDirectory finds all Go files in a directory recursively and replaces imports
func replaceImportsInDirectory(dir string) error {
	fmt.Printf("Replacing imports in %s\n", dir)

	// First make all files writable
	err := makeFilesWritable(dir)
	if err != nil {
		fmt.Printf("Warning: Could not make all files writable: %v\n", err)
		// Continue anyway
	}

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		fmt.Printf("Processing %s\n", path)

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Perform replacements
		modifiedContent := string(content)
		for _, rep := range replacements {
			if !strings.HasPrefix(rep.from, "//") {
				modifiedContent = strings.ReplaceAll(modifiedContent, rep.from, rep.to)
			}
		}

		// Write back to file if content changed
		if modifiedContent != string(content) {
			// First ensure we have permission to write to the file
			err = os.Chmod(path, 0644)
			if err != nil {
				fmt.Printf("Warning: Could not change file permissions: %v\n", err)
				// Continue anyway and try to write
			}

			err = os.WriteFile(path, []byte(modifiedContent), 0644)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// makeFilesWritable sets write permissions for all files in a directory
func makeFilesWritable(dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files, not directories
		if !info.IsDir() {
			// Add write permission to the file
			return os.Chmod(path, info.Mode()|0200) // Add write permission
		}
		return nil
	})
}
