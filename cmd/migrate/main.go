package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dir     string
		dsn     string
		command string
		steps   int
	)

	flag.StringVar(&dir, "dir", "database/migrations", "migrations directory")
	flag.StringVar(&dsn, "dsn", os.Getenv("NEXUS_DATABASE_URL"), "database DSN")
	flag.StringVar(&command, "cmd", "up", "command: up, down, steps, version, force")
	flag.IntVar(&steps, "steps", 0, "number of steps (for 'steps' command)")
	flag.Parse()

	if dsn == "" {
		log.Fatal("database DSN required: set -dsn or NEXUS_DATABASE_URL")
	}

	m, err := migrate.New("file://"+dir, dsn)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "steps":
		err = m.Steps(steps)
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			log.Fatal(verr)
		}
		fmt.Printf("version: %d, dirty: %v\n", v, dirty)
		return
	case "force":
		err = m.Force(steps)
	default:
		log.Fatalf("unknown command: %s", command)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration %s failed: %v", command, err)
	}
	fmt.Printf("migration %s completed\n", command)
}
