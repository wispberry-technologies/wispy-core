package tests

import (
	"fmt"

	"github.com/wispberry-technologies/wispy-core/common"
)

func TestPathSecurity() {
	// Test secure file reading
	fmt.Println("Testing secure file operations...")

	// Test 1: Normal file access should work
	fmt.Println("\n1. Testing normal file access...")
	exists := common.SecureExists("sites/example.com/config/config.toml")
	fmt.Printf("Config file exists: %v\n", exists)

	// Test 2: Path traversal attempts should be blocked
	fmt.Println("\n2. Testing path traversal protection...")

	// Try to access a file outside the site directory
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"../other-site/config/config.toml",
		"/etc/passwd",
		"C:\\Windows\\System32\\config\\sam",
	}

	for _, path := range maliciousPaths {
		exists := common.SecureExists(path)
		_, err := common.SecureReadFile(path)

		fmt.Printf("Path: %s\n", path)
		fmt.Printf("  Exists: %v\n", exists)
		fmt.Printf("  Read error: %v\n", err != nil)

		if exists || err == nil {
			fmt.Printf("  ⚠️  WARNING: Path traversal may not be properly blocked!\n")
		} else {
			fmt.Printf("  ✅ Path properly blocked\n")
		}
		fmt.Println()
	}

	// Test 3: Glob operations should be constrained
	fmt.Println("3. Testing secure glob operations...")

	// Normal glob should work
	files, err := common.SecureGlob("*.html")
	fmt.Printf("Normal glob (*.html): %d files, error: %v\n", len(files), err)

	// Malicious glob should be blocked
	files, err = common.SecureGlob("../../../*")
	fmt.Printf("Malicious glob (../../../*): %d files, error: %v\n", len(files), err)

	if err != nil {
		fmt.Println("  ✅ Malicious glob properly blocked")
	} else {
		fmt.Println("  ⚠️  WARNING: Malicious glob was not blocked!")
	}

	fmt.Println("\nSecurity test completed.")
}
