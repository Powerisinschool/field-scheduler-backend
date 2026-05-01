package config

import (
	"os"
)

type Config struct {
	ServerAddress    string
	DBUrl            string
	ParserServiceUrl string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		// Default local database URL for development
		dbURL = "postgres://postgres:password@localhost:5432/field_scheduler?sslmode=disable"
	}

	parserURL := os.Getenv("PARSER_SERVICE_URL")
	if parserURL == "" {
		parserURL = "http://localhost:8000"
	}

	return Config{
		ServerAddress:    ":" + port,
		DBUrl:            dbURL,
		ParserServiceUrl: parserURL,
	}
}
