package tests

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

	// Change to the tests directory
	testDir := "/home/theodore/Desktop/wispy-core/tests"
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("❌ Failed to change to test directory: %v\n", err)
		os.Exit(1)
	}

	// Build the go test command
	var cmd *exec.Cmd
	args := []string{"test"}

	// Add verbosity
	if *verbose {
		args = append(args, "-v")
	}

	// Add short mode
	if *short {
		args = append(args, "-short")
	}

	// Handle specific test
	if *testName != "" {
		args = append(args, "-run", *testName)
		fmt.Printf("🎯 Running specific test: %s\n", *testName)
	}

	// Handle test suites
	if *suite != "" {
		switch strings.ToLower(*suite) {
		case "all":
			fmt.Println("🏃 Running all tests...")
			args = append(args, "./...")
		case "api":
			fmt.Println("🔌 Running API tests...")
			args = append(args, "-run", "TestBasicAPIFunctionality|TestCachingPerformance|TestHTTPMethodHandling|TestConcurrentAccess|TestCacheStatistics")
		case "template":
			fmt.Println("📝 Running template function tests...")
			args = append(args, "-run", "TestTemplateFunctions|TestRenderEngineFunctionIntegration")
		case "comprehensive":
			fmt.Println("🔍 Running comprehensive tests...")
			args = append(args, "-run", "TestComprehensiveCaching|TestEdgeCases")
		case "performance":
			fmt.Println("⚡ Running performance tests...")
			args = append(args, "-run", "TestPerformanceBenchmark|TestMemoryAndResourceUsage")
		default:
			fmt.Printf("❌ Unknown test suite: %s\n", *suite)
			fmt.Println("Available suites: all, api, template, comprehensive, performance")
			os.Exit(1)
		}
	}

	// Handle benchmarks
	if *bench {
		fmt.Println("📊 Running benchmarks...")
		args = append(args, "-bench", ".")
		args = append(args, "-benchmem")
	}

	// If no specific test or suite specified, run all tests
	if *testName == "" && *suite == "" && !*bench {
		fmt.Println("🏃 Running all tests (default)...")
		args = append(args, "./...")
	}

	// Run the command
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🔧 Executing: go %s\n", strings.Join(args, " "))
	fmt.Println()

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Tests completed successfully!")
}

// Helper function to list available tests
func listTests() {
	fmt.Println("📋 Available tests:")

	testFiles, _ := filepath.Glob("*_test.go")
	for _, file := range testFiles {
		fmt.Printf("   📄 %s\n", file)
	}

	fmt.Println("\n🎯 Available test suites:")
	fmt.Println("   🔌 api        - Basic API functionality, caching, HTTP methods")
	fmt.Println("   📝 template   - Template function integration")
	fmt.Println("   🔍 comprehensive - Comprehensive caching and edge cases")
	fmt.Println("   ⚡ performance - Performance and memory tests")
	fmt.Println("   🏃 all        - All tests")

	fmt.Println("\n📊 Available benchmarks:")
	fmt.Println("   BenchmarkCachedVsUncached")
	fmt.Println("   BenchmarkConcurrentAccess")
}
