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

# CRITICAL: Explicitly copy the credentials file
COPY credentials.json /app/credentials.json

# Build the binary
RUN go build -o email-detector .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the binary from build stage
COPY --from=build /app/email-detector .

# Copy the credentials from the build stage
COPY --from=build /app/credentials.json .

# Set the environment variable for Firebase to find the credentials
ENV GOOGLE_APPLICATION_CREDENTIALS="/root/credentials.json"

# Run the binary
CMD ["./email-detector"]
