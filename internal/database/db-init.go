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
	// Prioritize loading .env from the executable's directory, then fall back to the current working directory, and finally use environment variables.
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			// .env exists beside the executable
			if err := godotenv.Load(envPath); err == nil {
				log.Printf("Loaded .env from executable directory: %s", envPath)
			} else {
				log.Printf("Found .env at %s but failed to load: %v", envPath, err)
			}
		} else {
			// Try loading .env from current working directory
			if err := godotenv.Load(); err == nil {
				log.Println("Loaded .env from current working directory")
			} else {
				log.Println("No .env file found, attempting to use environment variables")
			}
		}
	} else {
		// Could not determine executable path; try current working directory
		if err := godotenv.Load(); err == nil {
			log.Println("Loaded .env from current working directory")
		} else {
			log.Println("No .env file found, attempting to use environment variables")
		}
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
		configDir, err := getConfigDir()
		if err != nil {
			log.Fatalf("Failed to get configuration directory: %v", err)
		}
		dbPath := filepath.Join(configDir, "shipyard.db")
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Failed to open local database: %v", err)
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
