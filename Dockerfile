# Use the official Go image as the base image
FROM golang:1.23.2-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use lightweight alpine image for the final container
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the application port (modify according to your actual needs)
EXPOSE 8080

# Set environment variable for the database host, defaulting to 127.0.0.1 if not provided
ENV DB_HOST=127.0.0.1
ENV DB_PASSWORD=123456789

# Run the application
CMD ["./main"]
