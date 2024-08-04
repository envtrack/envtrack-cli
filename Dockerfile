# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .

# Define build arguments
ARG VERSION
ARG COMMIT_HASH
ARG BUILD_TIME_UTC

# Install git for getting the commit hash
RUN apk add --no-cache git

# Set default values for build arguments if not provided
RUN if [ -z "$VERSION" ]; then VERSION=$(git describe --tags --always --dirty); fi
RUN if [ -z "$COMMIT_HASH" ]; then COMMIT_HASH=$(git rev-parse HEAD); fi
RUN if [ -z "$BUILD_TIME_UTC" ]; then BUILD_TIME_UTC=$(date -u +"%Y-%m-%dT%H:%M:%SZ"); fi

RUN go mod download
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