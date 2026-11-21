# Stage 1: Build the Go application
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy application source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o employee-app .

# Stage 2: Create a lightweight runtime image
FROM alpine:3.22

WORKDIR /app

# Copy compiled binary from builder stage
COPY --from=builder /app/employee-app .

# Application port
EXPOSE 8080

# Start the application
CMD ["./employee-app"]
