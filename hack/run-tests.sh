#!/usr/bin/env bash

# Copyright 2024 Observatory Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "${REPO_ROOT}"

# Parse arguments
COVERAGE=false
VERBOSE=false
RUN_PATTERN=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--run)
            RUN_PATTERN="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-c|--coverage] [-v|--verbose] [-r|--run PATTERN]"
            exit 1
            ;;
    esac
done

echo "🧪 Running tests..."

# Setup test environment
export KUBEBUILDER_ASSETS=$(setup-envtest use --bin-dir /tmp/kubebuilder 1.28.0 -p path)

# Build test flags
TEST_FLAGS="-race"

if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [ -n "$RUN_PATTERN" ]; then
    TEST_FLAGS="$TEST_FLAGS -run $RUN_PATTERN"
fi

if [ "$COVERAGE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -coverprofile=coverage.out -covermode=atomic"
fi

# Run Go tests
echo "Running Go tests..."
go test $TEST_FLAGS ./api/... ./controllers/... ./cmd/...

if [ "$COVERAGE" = true ]; then
    echo ""
    echo "📊 Coverage Report:"
    go tool cover -func=coverage.out | tail -n 1
    echo ""
    echo "Generate HTML report: go tool cover -html=coverage.out -o coverage.html"
fi

# Run Python tests if they exist
if [ -d "tests" ] && [ -f "requirements.txt" ]; then
    echo ""
    echo "Running Python tests..."
    if command -v python3 &> /dev/null; then
        python3 -m pytest tests/ -v
    else
        echo "⚠️  Python3 not found, skipping Python tests"
    fi
fi

echo ""
echo "✅ All tests passed!"
