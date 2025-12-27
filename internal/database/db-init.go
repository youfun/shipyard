package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"youfun/shipyard/internal/crypto"
)

// ErrAppNotFound is a specific error returned when an application is not found.
var ErrAppNotFound = errors.New("application not found")

// ErrInstanceNotFound is a specific error returned when an application instance is not found.
var ErrInstanceNotFound = errors.New("application instance not found")

// ErrHostNotFound is a specific error returned when a host is not found.
var ErrHostNotFound = errors.New("host not found")

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed migrations_pg/*.sql
var migrationsPGFS embed.FS

var DB *sqlx.DB

// DBType represents the type of database being used
type DBType string

const (
	DBTypePostgres DBType = "postgres"
	DBTypeTurso    DBType = "libsql"
	DBTypeSQLite   DBType = "sqlite3"
)

// CurrentDBType holds the current database type
var CurrentDBType DBType

// Placeholder converts a query with ? placeholders to the appropriate format for the current database
// For PostgreSQL, it converts ? to $1, $2, etc.
// For SQLite/TURSO, it leaves the ? as-is
func Placeholder(query string) string {
	if CurrentDBType != DBTypePostgres {
		return query
	}
	// Convert ? placeholders to $1, $2, etc. for PostgreSQL
	var result strings.Builder
	paramIndex := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			result.WriteString(fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		} else {
			result.WriteByte(query[i])
		}
	}
	return result.String()
}

// Rebind rebinds a query to use the appropriate placeholder format for the current database
// This is a convenience wrapper around sqlx.DB.Rebind
func Rebind(query string) string {
	if DB == nil {
		return query
	}
	return DB.Rebind(query)
}

// getConfigDir gets or creates the ~/.shipyard configuration directory.
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".shipyard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// Priority: PostgreSQL > TURSO > Local SQLite
func InitDB() {
	// Prioritize loading .env from:
	// 1. Executable directory
	// 2. Current working directory
	// 3. /etc/shipyard/.env (Standard Linux install location)
	// 4. Environment variables (implicit)

	envLoaded := false
	var envPaths []string

	// 1. Executable directory
	if exePath, err := os.Executable(); err == nil {
		envPaths = append(envPaths, filepath.Join(filepath.Dir(exePath), ".env"))
	}

	// 2. Current working directory
	if wd, err := os.Getwd(); err == nil {
		envPaths = append(envPaths, filepath.Join(wd, ".env"))
	}

	// 3. Standard Linux configuration directory
	envPaths = append(envPaths, "/etc/shipyard/.env")

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				log.Printf("Loaded .env from: %s", path)
				envLoaded = true
				break
			} else {
				log.Printf("Found .env at %s but failed to load: %v", path, err)
			}
		}
	}

	if !envLoaded {
		log.Println("No .env file found in standard locations, attempting to use environment variables")
	}

	// Initialize Crypto with key from env or insecure default
	encryptionKey := os.Getenv("DEPLOYER_ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Println("⚠️  WARNING: DEPLOYER_ENCRYPTION_KEY not set. Using insecure default key for backward compatibility.")
		log.Println("Please set DEPLOYER_ENCRYPTION_KEY in your .env file to a secure 32-byte hex string.")
		// Default insecure key (formerly hardcoded in crypto.go)
		encryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	}
	if err := crypto.Init(encryptionKey); err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Priority: PostgreSQL > TURSO > Local SQLite
	pgURL := os.Getenv("DATABASE_URL")         // PostgreSQL connection string
	dbURL := os.Getenv("TURSO_DATABASE_URL")   // Turso database URL
	authToken := os.Getenv("TURSO_AUTH_TOKEN") // Turso auth token

	var db *sql.DB
	var err error
	var driverName string
	var migrateDialect string
	var migrationSource *migrate.EmbedFileSystemMigrationSource

	if pgURL != "" {
		// Using PostgreSQL
		log.Printf("Connecting to PostgreSQL database...")
		db, err = sql.Open("postgres", pgURL)
		if err != nil {
			log.Fatalf("Failed to open PostgreSQL database: %v", err)
		}
		driverName = "postgres"
		CurrentDBType = DBTypePostgres
		migrateDialect = "postgres"
		migrationSource = &migrate.EmbedFileSystemMigrationSource{
			FileSystem: migrationsPGFS,
			Root:       "migrations_pg",
		}
	} else if dbURL != "" {
		// Using Turso
		fullURL := dbURL
		if authToken != "" {
			fullURL = fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
		}
		log.Printf("Connecting to Turso database: %s", dbURL)
		db, err = sql.Open("libsql", fullURL)
		if err != nil {
			log.Fatalf("Failed to open Turso database: %v", err)
		}
		driverName = "libsql"
		CurrentDBType = DBTypeTurso
		migrateDialect = "sqlite3"
		migrationSource = &migrate.EmbedFileSystemMigrationSource{
			FileSystem: migrationsFS,
			Root:       "migrations",
		}
	} else {
		// Falling back to local SQLite
		log.Println("DATABASE_URL or TURSO_DATABASE_URL not configured, falling back to local SQLite")

		var dbPath string
		if envDBPath := os.Getenv("DB_PATH"); envDBPath != "" {
			dbPath = envDBPath
			// Ensure the directory exists
			if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
				log.Fatalf("Failed to create database directory: %v", err)
			}
			log.Printf("Using configured DB_PATH: %s", dbPath)
		} else {
			configDir, err := getConfigDir()
			if err != nil {
				log.Fatalf("Failed to get configuration directory: %v", err)
			}
			dbPath = filepath.Join(configDir, "shipyard.db")
		}

		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Failed to open local database: %v", err)
		}
		// Enable WAL mode for better concurrency and performance
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			log.Printf("Warning: Failed to enable WAL mode: %v", err)
		}

		driverName = "sqlite3"
		CurrentDBType = DBTypeSQLite
		migrateDialect = "sqlite3"
		migrationSource = &migrate.EmbedFileSystemMigrationSource{
			FileSystem: migrationsFS,
			Root:       "migrations",
		}
	}

	DB = sqlx.NewDb(db, driverName)

	if err = DB.Ping(); err != nil {
		log.Fatalf("Database Ping failed: %v", err)
	}

	log.Println("Running database migrations...")

	n, err := migrate.Exec(DB.DB, migrateDialect, migrationSource, migrate.Up)
	if err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	if n > 0 {
		log.Printf("Successfully applied %d new migrations!\n", n)
	} else {
		log.Println("Database initialization complete")
	}
}
