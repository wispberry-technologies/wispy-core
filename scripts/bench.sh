#!/usr/bin/env bash


echo "Benching..."

echo "Running benchmarks for template engine baseline..."
cd scripts/benchmarks/template-engine-baseline
go test -bench=. -run=^$ -benchmem -count=3 -v
echo "Benchmarks complete."
cd ../../..