#!/bin/bash
APP_NAME="generate_run_config"

# macOS
GOOS=darwin GOARCH=amd64 go build -o ${APP_NAME}-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o ${APP_NAME}-darwin-arm64 main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o ${APP_NAME}-linux-amd64 main.go
GOOS=linux GOARCH=arm64 go build -o ${APP_NAME}-linux-arm64 main.go

echo "Built binaries:"
ls -la ${APP_NAME}-*