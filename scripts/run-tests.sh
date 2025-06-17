#!/bin/bash

# Set environment variables for testing
export TEST_MODE=true
export SITES_PATH=${SITES_PATH:-$(pwd)/sites}
export ENV=${ENV:-test}
s
echo "Running tests with settings:"
echo "  - Environment: $ENV"
echo "  - Sites path: $SITES_PATH"

# Navigate to root directory
cd "$(dirname "$0")/.."

# Run all tests with verbose output
echo "Running all tests..."
go test -v ./...

# Exit with the same status code as the tests
exit $?
