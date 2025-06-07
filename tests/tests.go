package tests

func Run(testsToRun ...string) {

	// Map of test names to test functions
	testFuncs := map[string]func(){
		"path_security": TestPathSecurity,
		// Add more tests here as needed
	}

	if len(testsToRun) == 0 {
		// If no specific tests are requested, run all tests
		for name, testFunc := range testFuncs {
			println("Running test:", name)
			testFunc()
		}
	} else {
		// Run only the specified tests
		for _, testName := range testsToRun {
			if testFunc, exists := testFuncs[testName]; exists {
				println("Running test:", testName)
				testFunc()
			} else {
				println("Test not found:", testName)
			}
		}
	}

	// For now, just print a completion message
	println("All tests completed.")
}
