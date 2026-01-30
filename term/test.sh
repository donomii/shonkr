#!/bin/bash

echo "🧪 Testing Shonkr Terminal build..."

# Test compilation
echo "Testing compilation..."
go build -o test_shonkr shonkr_terminal_complete.go controller.go config.go

if [ $? -eq 0 ]; then
    echo "✅ Compilation successful!"
    ls -la test_shonkr
    
    # Clean up test binary
    rm -f test_shonkr
    
    echo ""
    echo "🎉 Your terminal is ready!"
    echo "Run with: go run shonkr_terminal_complete.go controller.go config.go"
    echo "Or build with: ./build.sh"
else
    echo "❌ Compilation failed"
    exit 1
fi
