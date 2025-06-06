package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var (
		testName = flag.String("test", "", "Specific test to run (e.g., TestBasicAPIFunctionality)")
		suite    = flag.String("suite", "", "Test suite to run (all, api, template, comprehensive, performance)")
		verbose  = flag.Bool("v", false, "Verbose output")
		bench    = flag.Bool("bench", false, "Run benchmarks")
		short    = flag.Bool("short", false, "Run tests in short mode")
	)
	flag.Parse()

	fmt.Println("🧪 Wispy Core Test Runner")
	fmt.Println("========================")

	// Get the parent directory of the runner (tests directory)
	workingDir, _ := os.Getwd()
	testDir := filepath.Join(workingDir, "..")
	if absTestDir, err := filepath.Abs(testDir); err == nil {
		testDir = absTestDir
	}

	// Change to the tests directory
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("❌ Failed to change to test directory: %v\n", err)
		os.Exit(1)
	}

	// Build the go test command
	var cmd *exec.Cmd
	args := []string{"test"}

	if *verbose {
		args = append(args, "-v")
	}

	if *short {
		args = append(args, "-short")
	}

	if *bench {
		if *testName != "" {
			args = append(args, "-bench", *testName)
		} else {
			args = append(args, "-bench", ".")
		}
	}

	// Handle specific test
	if *testName != "" && !*bench {
		args = append(args, "-run", *testName)
	}

	// Handle test suites
	if *suite != "" {
		switch strings.ToLower(*suite) {
		case "all":
			args = append(args, "./...")
		case "api":
			args = append(args, "-run", "TestBasicAPIFunctionality|TestAdvancedAPIFeatures|TestAPIPerformance")
		case "template":
			args = append(args, "-run", "TestTemplateFunction")
		case "comprehensive":
			args = append(args, "-run", "TestComprehensive")
		case "performance":
			args = append(args, "-run", "TestPerformance|Benchmark")
			if !*bench {
				args = append(args, "-bench", ".")
			}
		default:
			fmt.Printf("❌ Unknown test suite: %s\n", *suite)
			fmt.Println("Available suites: all, api, template, comprehensive, performance")
			os.Exit(1)
		}
	}

	// If no specific options, run all tests
	if *testName == "" && *suite == "" && !*bench {
		args = append(args, "./...")
	}

	fmt.Printf("🚀 Running: go %s\n", strings.Join(args, " "))
	fmt.Println("---")

	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n❌ Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ All tests completed successfully!")
}
