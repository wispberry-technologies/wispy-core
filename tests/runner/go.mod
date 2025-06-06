module github.com/wispberry-technologies/wispy-core/tests/runner

go 1.21

// Local development path - in production, this would use the published module
replace github.com/wispberry-technologies/wispy-core => ../../

require github.com/wispberry-technologies/wispy-core v0.0.0
