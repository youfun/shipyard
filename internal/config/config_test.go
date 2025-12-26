package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	os.Remove("shipyard.toml")
	LoadConfig("testapp", "shipyard.toml")

	// Check that app name was set
	if AppConfig.App != "testapp" {
		t.Errorf("expected App to be 'testapp', got '%s'", AppConfig.App)
	}

	// Check default KeepReleases
	if AppConfig.KeepReleases != 3 {
		t.Errorf("expected KeepReleases to be 3, got %d", AppConfig.KeepReleases)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	content := `
app = "myapp"
domains = ["example.com", "www.example.com"]
primary_domain = "example.com"
keep_releases = 5

[env]
  PHX_SERVER = true
`
	tmpfile, err := os.CreateTemp("", "deployer.*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Rename(tmpfile.Name(), "shipyard.toml"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("shipyard.toml")

	LoadConfig("", "shipyard.toml")

	if AppConfig.App != "myapp" {
		t.Errorf("expected App to be 'myapp', got '%s'", AppConfig.App)
	}

	if len(AppConfig.Domains) != 2 {
		t.Errorf("expected Domains length to be 2, got %d", len(AppConfig.Domains))
	}

	if AppConfig.PrimaryDomain != "example.com" {
		t.Errorf("expected PrimaryDomain to be 'example.com', got '%s'", AppConfig.PrimaryDomain)
	}

	if AppConfig.KeepReleases != 5 {
		t.Errorf("expected KeepReleases to be 5, got %d", AppConfig.KeepReleases)
	}
}

func TestReadConfigFile(t *testing.T) {
	content := `
app = "testapp2"
domains = ["test.com"]
`
	tmpfile, err := os.CreateTemp("", "deployer.*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := ReadConfigFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadConfigFile failed: %v", err)
	}

	if cfg.App != "testapp2" {
		t.Errorf("expected App to be 'testapp2', got '%s'", cfg.App)
	}

	if len(cfg.Domains) != 1 || cfg.Domains[0] != "test.com" {
		t.Errorf("expected Domains to be ['test.com'], got %v", cfg.Domains)
	}
}
