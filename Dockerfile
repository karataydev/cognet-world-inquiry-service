# Build stage
FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Add necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/cognet-world-inquiry-service .

# Copy .env file
COPY .env .

# Use non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./cognet-world-inquiry-service"]
