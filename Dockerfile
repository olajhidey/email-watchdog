# Build stage
FROM golang:1.25-alpine AS build

# Set environment variables for building
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o email-detector .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the binary from build stage
COPY --from=build /app/email-detector .

# Run the binary
CMD ["./email-detector"]
