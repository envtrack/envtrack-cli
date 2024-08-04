#!/bin/bash

VERSION=$(git describe --tags --always --abbrev=0)
COMMIT_HASH=$(git rev-parse HEAD)
BUILD_TIME_UTC=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

BINARY_NAME="envtrack"
PLATFORMS=("windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

for PLATFORM in "${PLATFORMS[@]}"; do
    OS="${PLATFORM%/*}"
    ARCH="${PLATFORM#*/}"
    
    if [ $OS = "windows" ]; then
        EXTENSION=".exe"
    else
        EXTENSION=""
    fi

    OUTPUT="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}${EXTENSION}"
    
    echo "Building for $OS/$ARCH..."
    GOOS=$OS GOARCH=$ARCH go build -ldflags="-X 'github.com/envtrack/envtrack-cli/internal/commands.Version=${VERSION}' -X 'github.com/envtrack/envtrack-cli/internal/commands.CommitHash=${COMMIT_HASH}' -X 'github.com/envtrack/envtrack-cli/internal/commands.LocalBuildTime=${BUILD_TIME_UTC}'" -o "./dist/${OUTPUT}" ./cmd/envtrack
    # EXAMPLE FOR darwin/arm64
    # GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'github.com/envtrack/envtrack-cli/internal/commands.Version=${VERSION}' -X 'github.com/envtrack/envtrack-cli/internal/commands.CommitHash=${COMMIT_HASH}' -X 'github.com/envtrack/envtrack-cli/internal/commands.LocalBuildTime=${BUILD_TIME_UTC}'" -o "./dist/${OUTPUT}" ./cmd/envtrack
    
    if [ $? -ne 0 ]; then
        echo "An error has occurred! Aborting the script execution..."
        exit 1
    fi
done