package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		databaseURL = flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection URL")
		migrations  = flag.String("migrations", envOr("MIGRATIONS_PATH", "migrations"), "Path to SQL migrations")
		direction   = flag.String("direction", "up", "Migration direction: up or down")
		steps       = flag.Int("steps", 0, "Number of migration steps (0 = all)")
	)
	flag.Parse()

	if *databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required (flag -database-url or env)")
		return 1
	}

	sourceURL := "file://" + *migrations
	m, err := migrate.New(sourceURL, *databaseURL)
	if err != nil {
		log.Printf("migrate init: %v", err)
		return 1
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migrate close source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migrate close database: %v", dbErr)
		}
	}()

	var runErr error
	switch *direction {
	case "up":
		if *steps > 0 {
			runErr = m.Steps(*steps)
		} else {
			runErr = m.Up()
		}
	case "down":
		if *steps > 0 {
			runErr = m.Steps(-*steps)
		} else {
			runErr = m.Down()
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown direction %q (use up or down)\n", *direction)
		return 1
	}

	if runErr != nil {
		if errors.Is(runErr, migrate.ErrNoChange) {
			log.Println("migrate: no change")
			return 0
		}
		log.Printf("migrate: %v", runErr)
		return 1
	}

	log.Println("migrate: ok")
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
