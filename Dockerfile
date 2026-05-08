# Stage 1: The Builder Environment
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to cache the module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your source code
COPY . .

# Replicate the `make build` command perfectly
RUN go build -o /app/bin/api cmd/api/main.go

# Stage 2: The Production Environment
FROM alpine:latest

WORKDIR /app

# Copy ONLY the compiled binary from the builder stage
COPY --from=builder /app/bin/api .

# Copy your migration files so the server has access to them
COPY --from=builder /app/internal/db/migration ./internal/db/migration

# Expose the port your Go app listens on (adjust if it is not 8080)
EXPOSE 8080

# Run the compiled binary
CMD ["./api"]
