# Start from the official Go image.
FROM golang:1.22-alpine

# Set the current working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
COPY go.mod ./
COPY go.sum ./

# Download dependencies.
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the Go application.
RUN go build -o email-detector .

# Expose the port your application listens on (if any).
# For a command-line tool, this might not be necessary.
# EXPOSE 8080 

# Command to run the executable.
CMD ["./email-detector"]
