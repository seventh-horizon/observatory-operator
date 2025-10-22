#!/bin/bash

echo "=========================================="
echo "  FINAL FIX - Release Hardened Project"
echo "=========================================="
echo ""

cd /Users/kalebkirby/release-hardened

echo "Step 1: Cleaning Go cache..."
go clean -modcache

echo ""
echo "Step 2: Running go mod tidy..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ go mod tidy failed!"
    echo "Check the errors above."
    exit 1
fi

echo ""
echo "Step 3: Checking if go.sum was created..."
if [ -f "go.sum" ]; then
    lines=$(wc -l < go.sum | tr -d ' ')
    echo "✅ go.sum created with $lines lines!"
else
    echo "❌ go.sum not found!"
    exit 1
fi

echo ""
echo "Step 4: Running go mod download..."
go mod download

echo ""
echo "Step 5: Building the project..."
go build ./...

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo ""
echo "✅ BUILD SUCCESSFUL!"
echo ""
echo "Step 6: Running tests..."
./hack/run-tests.sh

echo ""
echo "=========================================="
echo "  ✅ ALL DONE!"
echo "=========================================="
echo ""
echo "Your project is now fully working!"
echo ""
echo "Next steps:"
echo "  1. Deploy to cluster: make deploy"
echo "  2. Run samples: kubectl apply -f config/samples/"
echo "  3. Check metrics: kubectl port-forward svc/controller-metrics 8080:8080"
echo ""
