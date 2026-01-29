#!/bin/bash

# Prymis Browser Quick Run Script

# 1. Determine Go Binary
if ! command -v go &> /dev/null; then
    if [ -f "/usr/local/go/bin/go" ]; then
        GO_BIN="/usr/local/go/bin/go"
    else
        echo "❌ Error: Go is not installed."
        exit 1
    fi
else
    GO_BIN="go"
fi

# 2. Build quietly
$GO_BIN build -o prymis_gui ./cmd/prymis

# 3. Execution
if [ -f "./prymis_gui" ]; then
    ./prymis_gui
else
    echo "❌ Build failed."
    exit 1
fi
