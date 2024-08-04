# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Update go.sum and ensure all dependencies are downloaded
RUN go mod tidy
RUN go mod download all

# Define build arguments
ARG VERSION
ARG COMMIT_HASH
ARG BUILD_TIME_UTC

# Set default values for build arguments if not provided
RUN if [ -z "$VERSION" ]; then VERSION=$(git describe --tags --always --dirty); fi
RUN if [ -z "$COMMIT_HASH" ]; then COMMIT_HASH=$(git rev-parse HEAD); fi
RUN if [ -z "$BUILD_TIME_UTC" ]; then BUILD_TIME_UTC=$(date -u +"%Y-%m-%dT%H:%M:%SZ"); fi

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-X 'github.com/envtrack/envtrack-cli/internal/commands.Version=${VERSION}' \
    -X 'github.com/envtrack/envtrack-cli/internal/commands.CommitHash=${COMMIT_HASH}' \
    -X 'github.com/envtrack/envtrack-cli/internal/commands.LocalBuildTime=${BUILD_TIME_UTC}'" \
    -o envtrack ./cmd/envtrack

# Final stage
FROM alpine:3.17
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/envtrack .
ENTRYPOINT ["./envtrack"]