# Build stage
FROM golang:1.21-bookworm AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (needed for CGO and TigerBeetle)
RUN apt-get update && apt-get install -y git gcc libc6-dev && rm -rf /var/lib/apt/lists/*

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
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS requests and basic tools
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create a non-root user
RUN groupadd -g 1001 appuser && \
    useradd -r -u 1001 -g appuser appuser

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Change ownership to the non-root user
RUN chown appuser:appuser /app/main

# Switch to the non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["./main"]