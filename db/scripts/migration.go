package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().Str("script", "db-migration").Timestamp().Logger()

	// notify context of os.Interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// attach logger to context
	logger.WithContext(ctx)

	// Construct the database URL
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	// connect to DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to open database connection")
	}
	defer db.Close()

	// Create a new migrate instance with the PostgreSQL driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Panic().Err(err).Msg("failed to create migrations driver")
	}

	// Create a new migrate instance with the file source driver
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations", // Replace with the path to your migration files
		dbURL,
		driver,
	)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to create migration instance")
	}

	// Run the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Panic().Err(err).Msg("failed to run migrations")
	}

	fmt.Println("Migrations applied successfully")
}
