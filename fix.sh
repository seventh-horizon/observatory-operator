#!/bin/bash

echo "Fixing release-hardened project..."
echo ""

cd /Users/kalebkirby/release-hardened

echo "Step 1: Cleaning..."
go clean -modcache

echo ""
echo "Step 2: Running go mod tidy..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "ERROR: go mod tidy failed!"
    exit 1
fi

echo ""
echo "Step 3: Checking go.sum..."
if [ -f "go.sum" ]; then
    echo "SUCCESS: go.sum created!"
else
    echo "ERROR: go.sum not found!"
    exit 1
fi

echo ""
echo "Step 4: Building..."
go build ./...

if [ $? -ne 0 ]; then
    echo "ERROR: Build failed!"
    exit 1
fi

echo ""
echo "SUCCESS: Everything works!"
echo ""
echo "Next: Run tests with ./hack/run-tests.sh"
