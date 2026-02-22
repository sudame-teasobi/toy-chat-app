package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sudame/chat/pkg/env"
)

func main() {
	var (
		command = flag.String("cmd", "up", "Migration command: up, down, drop, version, force")
		steps   = flag.Int("steps", 0, "Number of migrations to apply (0 = all)")
		force   = flag.Int("force", -1, "Force set version (use with -cmd=force)")
	)
	flag.Parse()

	host := env.GetEnv("DB_HOST").WithDefault("localhost").Value()
	port := env.GetEnv("DB_PORT").WithDefault("4000").Value()
	user := env.GetEnv("DB_USER").WithDefault("root").Value()
	password := env.GetEnv("DB_PASSWORD").WithDefault("").Value()
	dbName := env.GetEnv("DB_NAME").WithDefault("toy_chat_app").Value()

	// TiDB compatibility settings:
	// - tidb_skip_isolation_level_check=1: Skip SERIALIZABLE isolation level check
	// - x-no-lock=true: Disable advisory locks
	// - multiStatements=true: Allow multiple SQL statements
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?tidb_skip_isolation_level_check=1&multiStatements=true&x-no-lock=true", user, password, host, port, dbName)

	m, err := migrate.New("file://sql/migrations", dsn)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer func() {
		if _, err := m.Close(); err != nil {
			log.Printf("Failed to close migrate instance: %v", err)
		}
	}()

	switch *command {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migration up completed successfully")

	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration down completed successfully")

	case "drop":
		err = m.Drop()
		if err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		fmt.Println("Drop completed successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if *force < 0 {
			log.Fatal("Please specify version with -force flag")
		}
		err = m.Force(*force)
		if err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("Forced version to %d\n", *force)

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}
