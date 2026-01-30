#!/bin/bash

echo "🚀 Building Shonkr Terminal..."

# Clean up any previous builds
rm -f shonkr_terminal

# Update dependencies
echo "📦 Updating dependencies..."
go mod tidy

# Build the complete terminal (all package files)
echo "🔨 Compiling..."
go build -o shonkr_terminal .

if [ -f "shonkr_terminal" ]; then
    echo "✅ Build successful!"
    echo "📊 Executable info:"
    ls -la shonkr_terminal
    echo ""
    echo "🎯 Ready to run:"
    echo "   ./shonkr_terminal"
    echo ""
else
    echo "❌ Build failed - executable not found"
    exit 1
fi
