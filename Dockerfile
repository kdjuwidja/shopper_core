# Use official Golang image as base
FROM golang:1.24.0

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go mod tidy && go build -o main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"]

