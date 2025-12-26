package deploy

import (
	"testing"
)

func TestShellEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple path",
			input: "/var/lib/app",
			want:  "'/var/lib/app'",
		},
		{
			name:  "Path with spaces",
			input: "/var/lib/my app",
			want:  "'/var/lib/my app'",
		},
		{
			name:  "Path with single quote",
			input: "/var/lib/app's data",
			want:  "'/var/lib/app'\"'\"'s data'",
		},
		{
			name:  "Path with special characters",
			input: "/var/lib/app$test",
			want:  "'/var/lib/app$test'",
		},
		{
			name:  "Path with semicolon",
			input: "/var/lib/app;rm -rf /",
			want:  "'/var/lib/app;rm -rf /'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellEscape(tt.input)
			if got != tt.want {
				t.Errorf("shellEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Absolute path",
			value: "/var/lib/app/data.db",
			want:  true,
		},
		{
			name:  "Relative path with ./",
			value: "./data/app.db",
			want:  true,
		},
		{
			name:  "Relative path with ../",
			value: "../data/app.db",
			want:  true,
		},
		{
			name:  "Path with multiple segments",
			value: "/var/www/app/data",
			want:  true,
		},
		{
			name:  "HTTP URL",
			value: "http://localhost:5432/db",
			want:  false,
		},
		{
			name:  "HTTPS URL",
			value: "https://example.com/api",
			want:  false,
		},
		{
			name:  "Postgres URL",
			value: "postgres://user:pass@localhost:5432/mydb",
			want:  false,
		},
		{
			name:  "Database connection string",
			value: "postgresql://user:password@host:5432/database",
			want:  false,
		},
		{
			name:  "Empty string",
			value: "",
			want:  false,
		},
		{
			name:  "Simple value",
			value: "true",
			want:  false,
		},
		{
			name:  "Number",
			value: "5432",
			want:  false,
		},
		{
			name:  "Host:Port",
			value: "localhost:5432",
			want:  false,
		},
		{
			name:  "Comma-separated domains",
			value: "example.com,www.example.com",
			want:  false,
		},
		{
			name:  "Path-like but has colon (probably connection string)",
			value: "localhost:5432/database",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLocalPath(tt.value)
			if got != tt.want {
				t.Errorf("isLocalPath(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestExtractPathsFromEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envs     map[string]interface{}
		wantKeys []string
	}{
		{
			name: "Database path",
			envs: map[string]interface{}{
				"DATABASE_PATH": "/var/lib/app/data.db",
				"PORT":          "4000",
				"LOG_LEVEL":     "info",
			},
			wantKeys: []string{"DATABASE_PATH"},
		},
		{
			name: "Multiple path variables",
			envs: map[string]interface{}{
				"DATABASE_PATH": "/var/lib/app/data.db",
				"LOG_DIR":       "/var/log/app",
				"CONFIG_FILE":   "/etc/app/config.toml",
				"PORT":          "4000",
			},
			wantKeys: []string{"DATABASE_PATH", "LOG_DIR", "CONFIG_FILE"},
		},
		{
			name: "URL should not be detected as path",
			envs: map[string]interface{}{
				"DATABASE_URL": "postgres://user:pass@localhost:5432/mydb",
				"API_URL":      "https://api.example.com",
			},
			wantKeys: []string{},
		},
		{
			name: "Mixed paths and URLs",
			envs: map[string]interface{}{
				"DATABASE_PATH": "/var/lib/app/data.db",
				"DATABASE_URL":  "postgres://user:pass@localhost:5432/mydb",
				"STORAGE_DIR":   "/var/storage",
			},
			wantKeys: []string{"DATABASE_PATH", "STORAGE_DIR"},
		},
		{
			name: "Path variable with non-path value",
			envs: map[string]interface{}{
				"DATABASE_PATH": "memory", // SQLite can use :memory:
				"LOG_DIR":       "/var/log/app",
			},
			wantKeys: []string{"LOG_DIR"},
		},
		{
			name: "Non-path variables",
			envs: map[string]interface{}{
				"PORT":       "4000",
				"PHX_SERVER": "true",
				"PHX_HOST":   "example.com,www.example.com",
			},
			wantKeys: []string{},
		},
		{
			name: "Variable with _DB suffix",
			envs: map[string]interface{}{
				"SQLITE_DB": "/data/app.db",
				"REDIS_DB":  "0", // Redis DB number
			},
			wantKeys: []string{"SQLITE_DB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPathsFromEnvVars(tt.envs)
			
			// Check if the expected keys are present
			if len(got) != len(tt.wantKeys) {
				t.Errorf("extractPathsFromEnvVars() returned %d paths, want %d", len(got), len(tt.wantKeys))
			}
			
			for _, key := range tt.wantKeys {
				if _, exists := got[key]; !exists {
					t.Errorf("extractPathsFromEnvVars() missing expected key %q", key)
				}
			}
			
			// Check that no unexpected keys are present
			for key := range got {
				found := false
				for _, wantKey := range tt.wantKeys {
					if key == wantKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("extractPathsFromEnvVars() returned unexpected key %q", key)
				}
			}
		})
	}
}
