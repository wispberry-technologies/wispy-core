# Wispy Core Test Runner

A micro-package for running the Wispy Core test suite with various options.

## Usage

From the test runner directory:
```bash
cd tests/runner
go build -o test-runner .
./test-runner [options]
```

## Options

- `-test <name>`: Run a specific test (e.g., `TestBasicAPIFunctionality`)
- `-suite <name>`: Run a test suite (all, api, template, comprehensive, performance)
- `-v`: Verbose output
- `-bench`: Run benchmarks
- `-short`: Run tests in short mode

## Test Suites

- **all**: Run all tests
- **api**: API functionality tests (`TestBasicAPIFunctionality`, `TestAdvancedAPIFeatures`, `TestAPIPerformance`)
- **template**: Template function tests (`TestTemplateFunction*`)
- **comprehensive**: Comprehensive caching and performance tests (`TestComprehensive*`)
- **performance**: Performance and benchmark tests (`TestPerformance*`, `Benchmark*`)

## Examples

```bash
# Run all tests with verbose output
./test-runner -suite all -v

# Run only API tests
./test-runner -suite api

# Run a specific test
./test-runner -test TestPageManagement -v

# Run performance benchmarks
./test-runner -suite performance -bench

# Run tests in short mode (faster, skips some intensive tests)
./test-runner -suite all -short
```

## Directory Structure

```
tests/
├── runner/                    # Test runner micro-package
│   ├── main.go               # Test runner implementation
│   └── README.md             # This file
├── *_test.go                 # Actual test files
└── old_tests/                # Legacy tests (build ignored)
```

## Notes

- The test runner automatically changes to the tests directory before executing
- Old tests in `old_tests/` directory are ignored during compilation
- All tests use the `tests` package and follow Go testing conventions
