# Build stage
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssg ./cmd/ssg

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S ssg && \
    adduser -u 1001 -S ssg -G ssg

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ssg .

# Change ownership to non-root user
RUN chown ssg:ssg /app/ssg

# Switch to non-root user
USER ssg

# Expose port (if needed for future web interface)
EXPOSE 8080

# Set the binary as entrypoint
ENTRYPOINT ["./ssg"]
