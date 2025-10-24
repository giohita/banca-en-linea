# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (needed for CGO and TigerBeetle)
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Tidy dependencies and generate go.sum
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and libc for CGO binaries
RUN apk --no-cache add ca-certificates libc6-compat

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown appuser:appgroup main

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]