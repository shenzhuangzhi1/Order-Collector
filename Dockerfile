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

# Use lightweight alpine image
FROM alpine:latest  

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose application port (modify according to actual situation)
EXPOSE 8080

# Run the application
CMD ["./main"]
