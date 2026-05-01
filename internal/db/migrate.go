package db

import (
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Required for the URL connection
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migration/*.sql
var MigrationFS embed.FS

// RunDBMigration connects to the database via URL and applies all pending migrations.
func RunDBMigration(dbURL string) {
	// 1. Tell golang-migrate to read from our embedded filesystem
	d, err := iofs.New(MigrationFS, "migration")
	if err != nil {
		log.Fatalf("Failed to create migration source: %v", err)
	}

	// 2. Initialize the migrate instance using the database URL
	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// 3. Run the 'Up' migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrate up: %v", err)
	}

	log.Println("Database migrated successfully!")
}
