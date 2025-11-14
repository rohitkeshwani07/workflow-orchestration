package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Get database configuration
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	var databaseURL string

	if driver == "postgres" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "workflow")
		password := getEnv("DB_PASSWORD", "workflow_password")
		dbname := getEnv("DB_NAME", "workflows")
		sslmode := getEnv("DB_SSLMODE", "disable")

		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	} else {
		dbPath := getEnv("DATABASE_PATH", "./data/workflows.db")
		databaseURL = fmt.Sprintf("sqlite3://%s", dbPath)
	}

	migrationsPath := "file://migrations"

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("✅ Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		log.Println("✅ Migrations rolled back successfully")

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Force command requires version argument")
		}
		version := os.Args[2]
		var v int
		fmt.Sscanf(version, "%d", &v)
		if err := m.Force(v); err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		log.Printf("✅ Forced database version to %d\n", v)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		if dirty {
			log.Printf("Current version: %d (dirty)\n", version)
		} else {
			log.Printf("Current version: %d\n", version)
		}

	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Create command requires migration name")
		}
		name := os.Args[2]
		log.Printf("To create a new migration, run:\n")
		log.Printf("  migrate create -ext sql -dir migrations -seq %s\n", name)
		log.Println("\nOr manually create files:")
		log.Printf("  migrations/NNNNNN_%s.up.sql\n", name)
		log.Printf("  migrations/NNNNNN_%s.down.sql\n", name)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Migration tool for workflow orchestration")
	fmt.Println("\nUsage:")
	fmt.Println("  migrate [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down            Rollback the last migration")
	fmt.Println("  force [version] Force set the migration version")
	fmt.Println("  version         Show current migration version")
	fmt.Println("  create [name]   Show instructions to create a new migration")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("  DB_DRIVER       Database driver (postgres or sqlite)")
	fmt.Println("  DB_HOST         PostgreSQL host")
	fmt.Println("  DB_PORT         PostgreSQL port")
	fmt.Println("  DB_USER         PostgreSQL user")
	fmt.Println("  DB_PASSWORD     PostgreSQL password")
	fmt.Println("  DB_NAME         PostgreSQL database name")
	fmt.Println("  DB_SSLMODE      PostgreSQL SSL mode")
	fmt.Println("  DATABASE_PATH   SQLite database file path")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
