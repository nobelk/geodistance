FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o server .

FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server ./

# The container will by default pass '/app' as the allowed directory if no other command line arguments are provided
ENTRYPOINT ["./server"]
CMD ["/app"]
