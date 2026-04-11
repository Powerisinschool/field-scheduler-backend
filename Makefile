DB_URL=postgresql://postgres:password@localhost:5432/field_scheduler?sslmode=disable

.PHONY: postgres postgres-down migrate-up migrate-down sqlc run build test

# ==============================================================================
# Docker / Database commands
# ==============================================================================

# Start the postgres container
postgres:
	docker compose up -d

# Stop the postgres container
postgres-down:
	docker compose down

# ==============================================================================
# Migration commands
# ==============================================================================

# Run database migrations up
migrate-up:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose up

# Rollback database migrations
migrate-down:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose down

migrate-drop:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose drop -f

# Helper to create a new migration file (Usage: make migrate-new name=add_users_table)
migrate-new:
	migrate create -ext sql -dir internal/db/migration -seq $(name)

# ==============================================================================
# Development commands
# ==============================================================================

# Generate sqlc Go code from SQL queries
sqlc:
	sqlc generate

# Generate Swagger docs
swagger:
	swag init -g cmd/api/main.go

# Run the Go API server
run:
	go run cmd/api/main.go

# Build the Go API server binary
build:
	go build -o bin/api cmd/api/main.go

# Clean up generated binary
clean:
	rm -rf bin/api

# Clean up generated files, binaries, and build artifacts
clean-all:
	rm -rf bin/ internal/db/sqlc/ internal/api/docs/

# Run tests
test:
	go test -v ./...
