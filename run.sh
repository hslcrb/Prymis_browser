#!/bin/bash

# Prymis Browser Quick Run Script

echo "🚀 Starting Prymis Browser Engine..."

# 1. Check if Go is installed
if ! command -v go &> /dev/null; then
    # Try common local path if not in PATH
    if [ -f "/usr/local/go/bin/go" ]; then
        GO_BIN="/usr/local/go/bin/go"
    else
        echo "❌ Error: Go is not installed or not in PATH."
        exit 1
    fi
else
    GO_BIN="go"
fi

# 2. Build or Run
echo "📦 Compiling and Running..."
$GO_BIN run ./cmd/prymis

# 3. Check result
if [ $? -eq 0 ]; then
    echo "✅ Successfully rendered Prymis output!"
    if [ -f "output.png" ]; then
        echo "🖼️  Rendering saved to: output.png"
    fi
else
    echo "❌ Execution failed."
    exit 1
fi
