package database

import (
	"youfun/shipyard/internal/models"
	"os"
	"testing"

	"github.com/google/uuid"
)

// TestMain is used to set up and tear down the environment before and after tests.
func TestMain(m *testing.M) {
	// Use a temporary directory for testing.
	tmpDir, err := os.MkdirTemp("", "deployer-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set a temporary HOME directory.
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	InitDB()
	code := m.Run()
	DB.Close()
	os.Exit(code)
}

func TestAddAndGetSSHHost(t *testing.T) {
	password := "password"
	host := &models.SSHHost{
		ID:       uuid.New(),
		Name:     "test-host",
		Addr:     "localhost",
		Port:     22,
		User:     "tester",
		Password: &password,
	}

	err := AddSSHHost(host)
	if err != nil {
		t.Fatalf("AddSSHHost() failed: %v", err)
	}

	retrievedHost, err := GetSSHHostByName("test-host")
	if err != nil {
		t.Fatalf("GetSSHHostByName() failed: %v", err)
	}

	if retrievedHost == nil {
		t.Fatal("expected to find host, but returned nil")
	}

	if retrievedHost.Name != host.Name {
		t.Errorf("expected host name %s, but got %s", host.Name, retrievedHost.Name)
	}
}

func TestPlaceholder(t *testing.T) {
	// Test the placeholder function with simulated SQLite type
	originalDBType := CurrentDBType
	defer func() { CurrentDBType = originalDBType }()

	// Test SQLite mode - should leave ? as-is
	CurrentDBType = DBTypeSQLite
	query := Placeholder("SELECT * FROM test WHERE a = ? AND b = ?")
	expected := "SELECT * FROM test WHERE a = ? AND b = ?"
	if query != expected {
		t.Errorf("SQLite placeholder: expected %q, got %q", expected, query)
	}

	// Test PostgreSQL mode - should convert ? to $1, $2, etc.
	CurrentDBType = DBTypePostgres
	query = Placeholder("SELECT * FROM test WHERE a = ? AND b = ?")
	expected = "SELECT * FROM test WHERE a = $1 AND b = $2"
	if query != expected {
		t.Errorf("PostgreSQL placeholder: expected %q, got %q", expected, query)
	}

	// Test TURSO mode - should leave ? as-is (like SQLite)
	CurrentDBType = DBTypeTurso
	query = Placeholder("SELECT * FROM test WHERE a = ? AND b = ?")
	expected = "SELECT * FROM test WHERE a = ? AND b = ?"
	if query != expected {
		t.Errorf("TURSO placeholder: expected %q, got %q", expected, query)
	}
}

func TestDBType(t *testing.T) {
	// Test that the CurrentDBType is set correctly during initialization
	// Since we're using SQLite in tests (no env vars set), it should be sqlite3
	if CurrentDBType != DBTypeSQLite {
		t.Errorf("Expected DBTypeSQLite, got %s", CurrentDBType)
	}
}
