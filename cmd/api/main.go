package main

import (
	"context"
	"log"
	"net/http"

	"field-scheduler-backend/internal/api"
	"field-scheduler-backend/internal/api/handlers"
	"field-scheduler-backend/internal/config"
	"field-scheduler-backend/internal/core/services"
	dbx "field-scheduler-backend/internal/db"
	db "field-scheduler-backend/internal/db/sqlc"
	python_api "field-scheduler-backend/internal/infrastructure/python_api"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Setup Database Connection Pool
	ctx := context.Background()

	log.Println("Running database migrations...")
	dbx.RunDBMigration(cfg.DBUrl)

	connPool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer connPool.Close()

	// 3. Initialize Repositories (sqlc)
	repo := db.New(connPool)

	// 4. Initialize the Python Parser Client
	rawPythonClient := python_api.Client{BaseURL: cfg.ParserServiceUrl, HTTPClient: &http.Client{}}

	// 5. Initialize Services
	parserService := services.NewParserService(&rawPythonClient, repo)
	conductorService := services.NewConductorService(repo)
	venueService := services.NewVenueService(repo)
	scheduleService := services.NewScheduleService(repo)
	blockService := services.NewBlockService(repo)

	// 6. Initialize Handlers
	conductorHandler := handlers.NewConductorHandler(conductorService)
	venueHandler := handlers.NewVenueHandler(venueService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService, conductorService, parserService, venueService)
	mapHandler := handlers.NewMapHandler(parserService, blockService)

	// 7. Setup Router
	router := api.SetupRouter(cfg.SwaggerHost, scheduleHandler, conductorHandler, venueHandler, mapHandler)

	// 8. Start Server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}
