# Field Scheduler Backend

A REST API for managing field schedules, conductors, venues, and geographic territory maps. Built with Go and backed by PostgreSQL, it integrates with a Python microservice for PDF parsing and geospatial data processing.

## Features

- Schedule management — create, list, and batch-update field service entries by date range
- PDF upload — parse and sync schedule data from uploaded PDF files
- Conductor and venue management
- Geographic territory mapping — create and manage cards (regions) and blocks (territories) with GeoJSON coordinate support
- OSMnx-generated map blocks via the Python data processing service
- CSV bulk import for blocks and JSON-based data import/restore
- Auto-generated Swagger/OpenAPI documentation

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP | Standard library `net/http` |
| Database | PostgreSQL 18 |
| DB Driver | `jackc/pgx/v5` |
| Query codegen | SQLC v1.30 |
| Migrations | `golang-migrate/migrate/v4` |
| API docs | Swaggo (Swagger UI) |
| Parser service | Python 3.11 + FastAPI |
| Containerisation | Docker + Docker Compose |

## Project Structure

```
field-scheduler-backend/
├── cmd/api/main.go                     # Entry point
├── internal/
│   ├── api/
│   │   ├── router.go                   # Route definitions
│   │   └── handlers/                   # HTTP handlers
│   ├── config/config.go                # Environment configuration
│   ├── core/
│   │   ├── models/                     # Domain models
│   │   └── services/                   # Business logic
│   ├── db/
│   │   ├── migration/                  # SQL migration files
│   │   ├── query/                      # SQLC query files
│   │   └── sqlc/                       # Generated type-safe DB code
│   └── infrastructure/
│       └── python_api/client.go        # HTTP client for parser service
├── data-processing-service/            # Python microservice (PDF + geospatial)
├── docs/                               # Generated Swagger docs
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── sqlc.yaml
```

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- `golang-migrate` CLI (for running migrations outside Docker)
- `sqlc` CLI (only needed to regenerate DB code)
- `swag` CLI (only needed to regenerate Swagger docs)

## Getting Started

### 1. Start infrastructure

```bash
make postgres
```

This starts a PostgreSQL 18 container on port `5432`.

### 2. Run migrations

```bash
make migrate-up
```

### 3. Configure environment

Copy and edit the example env file:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | API server port |
| `DB_URL` | `postgres://postgres:password@localhost:5432/field_scheduler?sslmode=disable` | PostgreSQL connection string |
| `PARSER_SERVICE_URL` | `http://localhost:8000` | Python data processing service URL |
| `SWAGGER_HOST` | `localhost:8080` | Host shown in Swagger UI |

### 4. Run the API

```bash
make run
```

The server starts at `http://localhost:8080`.

Swagger UI is available at `http://localhost:8080/api/docs/`.

## Docker Compose

To run the full stack (API, PostgreSQL, and Python parser service) together:

```bash
docker compose up
```

## API Endpoints

All routes are prefixed with `/api/v1`.

### Schedules

| Method | Path | Description |
|---|---|---|
| `GET` | `/schedules/` | List entries by date range (`start_date`, `end_date` query params) |
| `POST` | `/schedules/` | Create a schedule entry |
| `PUT` | `/schedules/` | Batch update schedule entries |
| `POST` | `/schedules/upload/` | Upload a PDF to parse and sync entries |

### Reference Data

| Method | Path | Description |
|---|---|---|
| `GET` | `/conductors/` | List all conductors |
| `GET` | `/venues/` | List all venues |

### Maps

| Method | Path | Description |
|---|---|---|
| `GET` | `/maps/generated-blocks/` | Fetch OSMnx-generated territory blocks |
| `GET` | `/maps/cards/` | List user-defined cards |
| `POST` | `/maps/cards/` | Create a card |
| `DELETE` | `/maps/cards/{id}/` | Delete a card |
| `GET` | `/maps/blocks/` | List user-defined blocks |
| `POST` | `/maps/blocks/` | Create a block |
| `DELETE` | `/maps/blocks/{id}/` | Delete a block |
| `DELETE` | `/maps/cards/{id}/blocks/` | Delete all blocks belonging to a card |
| `POST` | `/maps/blocks/upload/` | Upload a CSV to sync blocks |
| `POST` | `/maps/import/` | Import cards and blocks from JSON |

### Utility

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/docs/` | Swagger UI |

## Makefile Reference

```bash
make run            # Run the API server
make build          # Compile binary to ./bin/
make test           # Run tests
make migrate-up     # Apply all pending migrations
make migrate-down   # Roll back the last migration
make postgres       # Start PostgreSQL container
make postgres-down  # Stop PostgreSQL container
make sqlc           # Regenerate SQLC code from SQL queries
make swagger        # Regenerate Swagger docs
```

## Data Processing Service

The Python microservice lives in `data-processing-service/` and exposes:

- `POST /parse/pdf` — OCR-based PDF schedule parsing (OpenCV + Tesseract)
- `GET /fetch/map-blocks` — OSMnx geospatial block generation
- `POST /fetch/block-by-topology` — Block coordinate lookup by street topology

Map block results are cached for 14 days to reduce repeated OSMnx queries.

## Database Schema

Core tables:

- **users** — authentication accounts
- **conductors** — field workers linked to users
- **venues** — activity locations
- **events** — event definitions with date ranges
- **schedule_entries** — dated schedule records with conductor and venue assignments
- **cards** — named geographic regions with zoom and centre coordinates
- **blocks** — geographic polygons or linestrings belonging to a card (coordinates stored as JSONB)
- **schedule_entry_blocks** — junction table linking schedule entries to blocks
