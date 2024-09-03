# Start from a minimal Go image
FROM golang:1.23-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to cache dependencies downloading
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Copy the .env file to the /app directory
COPY .env /app/.env

# Build the Go application
RUN go build -o main ./cmd/auth_service/main.go


# Start a new stage from scratch
FROM alpine:latest

# Set the working directory for the runtime container
WORKDIR /app

# Copy the binary built in the previous stage
COPY --from=builder /app/main .

# Expose port 8080 (adjust according to your application's port)
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
