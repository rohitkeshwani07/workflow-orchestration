# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -tags "libsqlite3 sqlite_omit_load_extension" -o workflow-server main.go

# Build migration tool
RUN CGO_ENABLED=1 GOOS=linux go build -tags "libsqlite3 sqlite_omit_load_extension" -o migrate cmd/migrate/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binaries from builder
COPY --from=builder /app/workflow-server .
COPY --from=builder /app/migrate .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Create data directory
RUN mkdir -p /root/data

# Expose port
EXPOSE 3001

# Run the application
CMD ["./workflow-server"]
