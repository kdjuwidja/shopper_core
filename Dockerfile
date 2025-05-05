# Use official Golang image as base
FROM golang:1.24.0 AS go-builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN GOOS=linux go build -o main .

FROM golang:1.24.0

WORKDIR /app
# Copy the built go binary
COPY --from=go-builder /app/main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"]

